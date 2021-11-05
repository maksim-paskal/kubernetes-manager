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
)

const (
	ScaleDownUtcHourMinPeriod = 16
	ScaleDownUtcHourMaxPeriod = 22
)

func Schedule() {
	tracer := opentracing.GlobalTracer()

	for {
		<-time.After(*config.Get().BatchShedulePeriod)

		span := tracer.StartSpan("scheduleBatch")

		utcHour := time.Now().UTC().Hour()

		if utcHour >= ScaleDownUtcHourMinPeriod && utcHour <= ScaleDownUtcHourMaxPeriod {
			if err := scaleDownALL(span); err != nil {
				log.WithError(err).Error()
			}
		}

		if err := Execute(span); err != nil {
			log.WithError(err).Error()
		}

		span.Finish()
	}
}

func scaleDownALL(rootSpan opentracing.Span) error {
	tracer := opentracing.GlobalTracer()
	span := tracer.StartSpan("scaleDownALL", opentracing.ChildOf(rootSpan.Context()))

	defer span.Finish()

	ingresses, err := api.GetIngress()
	if err != nil {
		return errors.Wrap(err, "error listing ingresses")
	}

	for _, ingress := range ingresses {
		go func(ingress api.GetIngressList) {
			log := log.WithField("namespace", ingress.Namespace)

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

	git, err := gitlab.NewClient(*config.Get().GitlabToken, gitlab.WithBaseURL(*config.Get().GitlabURL))
	if err != nil {
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			Error()
	}

	ingresss, err := api.GetIngress()
	if err != nil {
		return errors.Wrap(err, "error list ingress")
	}

	for _, ingress := range ingresss {
		gitBranch := ingress.IngressAnotations[config.LabelGitBranch]
		gitProjectID := ingress.IngressAnotations[config.LabelGitProjectID]

		log := log.WithFields(log.Fields{
			"gitProjectID": gitProjectID,
			"gitBranch":    gitBranch,
		})

		isDeleteBranch := false

		_, _, err = git.Branches.GetBranch(gitProjectID, gitBranch)

		//nolint:gocritic
		if utils.IsSystemBranch(gitBranch) {
			isDeleteBranch = false

			log.WithField("isDeleteBranch", isDeleteBranch).Debug("is system branch")
		} else if err != nil {
			if strings.Contains(err.Error(), "404 Branch Not Found") {
				isDeleteBranch = true

				log.WithField("isDeleteBranch", isDeleteBranch).Debug("git branch not found")
			}
		} else if ingress.NamespaceLastScaledDays > *config.Get().RemoveBranchLastScaleDate {
			isDeleteBranch = true
		} else if ingress.NamespaceCreatedDays > 1 {
			isDeleteBranch = getLastCommitBranch(span, git, gitProjectID, gitBranch)

			log.WithField("isDeleteBranch", isDeleteBranch).Debug("namespace.CreationTimestamp.Time > 1h")
		}

		log.Debugf("isDeleteBranch=%t", isDeleteBranch)

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
