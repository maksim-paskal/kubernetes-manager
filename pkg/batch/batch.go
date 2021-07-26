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
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/api"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	logrushookopentracing "github.com/maksim-paskal/logrus-hook-opentracing"
	opentracing "github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Schedule() {
	duration, err := time.ParseDuration("30m")
	if err != nil {
		log.WithError(err).Fatal()
	}

	tracer := opentracing.GlobalTracer()

	for {
		<-time.After(duration)
		span := tracer.StartSpan("scheduleBatch") //nolint:wsl
		Execute(span)
		span.Finish()
	}
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

func Execute(rootSpan opentracing.Span) {
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

	opt := metav1.ListOptions{
		LabelSelector: *config.Get().IngressFilter,
	}

	ingresss, _ := api.Clientset.ExtensionsV1beta1().Ingresses("").List(context.TODO(), opt)

	for _, ingress := range ingresss.Items {
		gitBranch := ingress.Annotations[config.LabelGitBranch]
		gitProjectID := ingress.Annotations[config.LabelGitProjectID]

		log := log.WithFields(log.Fields{
			"gitProjectID": gitProjectID,
			"gitBranch":    gitBranch,
		})

		namespace, err := api.Clientset.CoreV1().Namespaces().Get(context.TODO(), ingress.Namespace, metav1.GetOptions{})
		if err != nil {
			log.
				WithError(err).
				WithField(logrushookopentracing.SpanKey, span).
				Error()

			return
		}

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
		} else if len(namespace.GetAnnotations()[config.LabelLastScaleDate]) > 0 {
			lastScaleDate, err := time.Parse(time.RFC3339, namespace.GetAnnotations()[config.LabelLastScaleDate])
			if err != nil {
				log.
					WithError(err).
					WithField(logrushookopentracing.SpanKey, span).
					Warn()
			}
			if utils.DiffToNow(lastScaleDate) > *config.Get().RemoveBranchLastScaleDate {
				isDeleteBranch = true

				log.WithField("isDeleteBranch", isDeleteBranch).Debug("lastScaleDate > removeBranchLastScaleDate")
			}
		} else if utils.DiffToNow(namespace.CreationTimestamp.Time) > 1 {
			isDeleteBranch = getLastCommitBranch(span, git, gitProjectID, gitBranch)

			log.WithField("isDeleteBranch", isDeleteBranch).Debug("namespace.CreationTimestamp.Time > 1h")
		}

		log.Debugf("isDeleteBranch=%t", isDeleteBranch)

		if isDeleteBranch {
			span.LogKV("delete branch", gitBranch)

			ch1 := make(chan api.HTTPResponse)
			q := make(url.Values)

			q.Add("namespace", ingress.Namespace)

			for k, v := range ingress.Annotations {
				if strings.HasPrefix(k, "kubernetes-manager") {
					q.Add(k[19:], v)
				}
			}

			go api.MakeAPICall(span, "/api/deleteALL", q, ch1)

			span.LogKV("result", <-ch1)
		}
	}
}
