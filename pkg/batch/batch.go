/*
Copyright paskal.maksim@gmail.com
Licensed under the Apache License, Version 2.0 (the "License")
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package batch

import (
	"fmt"
	"strings"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/api"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	logrushookopentracing "github.com/maksim-paskal/logrus-hook-opentracing"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/atomic"
)

var isStoped = *atomic.NewBool(false)

func Schedule() {
	log.Info("starting batch")

	isStoped.Store(false)

	tracer := opentracing.GlobalTracer()

	_, err := time.LoadLocation(*config.Get().BatchSheduleTimezone)
	if err != nil {
		log.WithError(err).Fatal()
	}

	for {
		<-time.After(*config.Get().BatchShedulePeriod)

		if isStoped.Load() {
			return
		}

		span := tracer.StartSpan("scheduleBatch")

		if err := Execute(span); err != nil {
			log.WithError(err).Error()
		}

		span.Finish()
	}
}

func Stop() {
	isStoped.Store(true)
}

func IsScaleDownActive(now time.Time) bool {
	batchSheduleTimezone, err := time.LoadLocation(*config.Get().BatchSheduleTimezone)
	if err != nil {
		log.WithError(err).Fatal()
	}

	timeMin := time.Date(now.Year(), now.Month(), now.Day(), config.ScaleDownHourMinPeriod, 0, 0, 0, batchSheduleTimezone)
	timeMax := time.Date(now.Year(), now.Month(), now.Day(), config.ScaleDownHourMaxPeriod, 0, 0, 0, batchSheduleTimezone)

	if now.After(timeMin) || now.Equal(timeMin) {
		return true
	}

	if now.Before(timeMax) || now.Equal(timeMax) {
		return true
	}

	return false
}

func scaleDownALL(rootSpan opentracing.Span) error {
	tracer := opentracing.GlobalTracer()
	span := tracer.StartSpan("scaleDownALL", opentracing.ChildOf(rootSpan.Context()))

	defer span.Finish()

	if !IsScaleDownActive(time.Now()) {
		log.Debug("scaleDownALL not in period")

		return nil
	}

	ingresses, err := api.GetIngress()
	if err != nil {
		return errors.Wrap(err, "error listing ingresses")
	}

	for _, ingress := range ingresses {
		go func(ingress *api.GetIngressList) {
			log := log.WithField("namespace", ingress.Namespace)

			scaleDelayText := ingress.NamespaceAnotations[config.LabelScaleDownDelay]
			if len(scaleDelayText) > 0 {
				scaleDelayTime, err := time.Parse(time.RFC3339, scaleDelayText)
				if err != nil {
					log.WithError(err).Error(err)
				} else if time.Now().Before(scaleDelayTime) {
					log.Info("scale down delay is active")
					// do not scale down if delay is active
					return
				}
			}

			log.Info("scaledown")

			err := api.ScaleNamespace(ingress.Namespace, 0)
			if err != nil {
				log.WithError(err).Error()
			}
		}(ingress)
	}

	return nil
}

func getLastCommitBranch(rootSpan opentracing.Span, git *gitlab.Client, gitProjectID string, gitBranch string) bool {
	tracer := opentracing.GlobalTracer()
	span := tracer.StartSpan("getLastCommitBranch", opentracing.ChildOf(rootSpan.Context()))

	defer span.Finish()

	lastCommitDateForRemove := time.Now().AddDate(0, 0, -*config.Get().RemoveBranchDaysInactive)

	gitCommitOptions := gitlab.ListCommitsOptions{
		RefName: &gitBranch,
		Since:   &lastCommitDateForRemove,
	}

	gitCommits, _, err := git.Commits.ListCommits(gitProjectID, &gitCommitOptions)
	if err != nil {
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			Error()
	}

	gitCommitsLen := len(gitCommits)

	if gitCommitsLen == 0 {
		log.Debugf("deleteOnLastCommit=gitBranch=%s,gitProjectID=%s", gitBranch, gitProjectID)

		return true
	}

	return false
}

func Execute(rootSpan opentracing.Span) error {
	tracer := opentracing.GlobalTracer()
	span := tracer.StartSpan("batch", opentracing.ChildOf(rootSpan.Context()))

	defer span.Finish()

	if err := scaleDownALL(span); err != nil {
		log.WithError(err).Error()
	}

	git := api.GetGitlabClient()
	if git == nil {
		return errors.New("no gitlab client")
	}

	ingresss, err := api.GetIngress()
	if err != nil {
		return errors.Wrap(err, "error list ingress")
	}

	for _, ingress := range ingresss {
		gitBranch := ingress.IngressAnotations[config.LabelGitBranch]
		gitProjectID := ingress.IngressAnotations[config.LabelGitProjectID]

		log := log.WithFields(log.Fields{
			"namespace":    ingress.Namespace,
			"gitProjectID": gitProjectID,
			"gitBranch":    gitBranch,
		})

		if utils.IsSystemNamespace(ingress.NamespaceName) {
			log.Debugf("%s is system namespace", ingress.NamespaceName)

			continue
		}

		if utils.IsSystemBranch(gitBranch) {
			log.Debugf("%s is system branch", gitBranch)

			continue
		}

		isDeleteBranch := false
		deleteReason := "branch will not deleted"

		_, _, err = git.Branches.GetBranch(gitProjectID, gitBranch)

		//nolint:gocritic
		if err != nil {
			if strings.Contains(err.Error(), "404 Branch Not Found") {
				isDeleteBranch = true
				deleteReason = "git branch not found"
			}
		} else if ingress.NamespaceLastScaledDays > *config.Get().RemoveBranchLastScaleDate {
			isDeleteBranch = true
			deleteReason = fmt.Sprintf("ingress.NamespaceLastScaledDays > %d", *config.Get().RemoveBranchLastScaleDate)
		} else if ingress.NamespaceCreatedDays > 1 {
			isDeleteBranch = getLastCommitBranch(span, git, gitProjectID, gitBranch)
			deleteReason = "namespace.NamespaceCreatedDays > 1"
		}

		log.WithField("isDeleteBranch", isDeleteBranch).Debug(deleteReason)

		if isDeleteBranch {
			deleteALLResult := api.DeleteALL(
				ingress.Namespace,
				ingress.IngressAnotations[config.LabelRegistryTag],
				ingress.IngressAnotations[config.LabelGitProjectID],
			)

			log.Info(deleteALLResult.JSON())
		}
	}

	return nil
}