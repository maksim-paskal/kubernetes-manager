package main

import (
	"net/url"
	"strings"
	"time"

	sentry "github.com/getsentry/sentry-go"
	opentracing "github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func scheduleBatch() {
	duration, err := time.ParseDuration("30m")
	if err != nil {
		sentry.CaptureException(err)
		sentry.Flush(time.Second)

		log.Panic(err)
	}

	var tracer = opentracing.GlobalTracer()

	for {
		<-time.After(duration)
		span := tracer.StartSpan("scheduleBatch")
		batch(span)
		span.Finish()
	}
}

func getLastCommitBranch(rootSpan opentracing.Span, git *gitlab.Client, gitProjectID string, gitBranch string) bool {
	var tracer = opentracing.GlobalTracer()
	span := tracer.StartSpan("getLastCommitBranch", opentracing.ChildOf(rootSpan.Context()))
	defer span.Finish()
	lastCommitDateForRemove := time.Now().AddDate(0, 0, -*appConfig.removeBranchDaysInactive)

	gitCommitOptions := gitlab.ListCommitsOptions{
		RefName: &gitBranch,
		Since:   &lastCommitDateForRemove,
	}

	gitCommits, _, err := git.Commits.ListCommits(gitProjectID, &gitCommitOptions)

	if err != nil {
		log.Error(err)
		logError(span, sentry.LevelInfo, nil, nil, err.Error())
	}
	if len(gitCommits) == 0 {
		log.Debugf("deleteOnLastCommit=gitBranch=%s,gitProjectID=%s", gitBranch, gitProjectID)
		return true
	}
	return false
}

func batch(rootSpan opentracing.Span) {
	var tracer = opentracing.GlobalTracer()
	span := tracer.StartSpan("batch", opentracing.ChildOf(rootSpan.Context()))
	defer span.Finish()

	git, err := gitlab.NewClient(*appConfig.gitlabToken, gitlab.WithBaseURL(*appConfig.gitlabURL))

	if err != nil {
		log.Error(err)
		logError(span, sentry.LevelInfo, nil, nil, err.Error())
	}

	opt := metav1.ListOptions{
		LabelSelector: *appConfig.ingressFilter,
	}

	ingresss, _ := clientset.ExtensionsV1beta1().Ingresses("").List(opt)

	for _, ingress := range ingresss.Items {
		gitBranch := ingress.Annotations[label_gitBranch]
		gitProjectID := ingress.Annotations[label_gitProjectId]

		namespace, err := clientset.CoreV1().Namespaces().Get(ingress.Namespace, metav1.GetOptions{})
		if err != nil {
			log.Error(err)
			logError(span, sentry.LevelError, nil, err, "")
			return
		}

		isDeleteBranch := false

		_, _, err = git.Branches.GetBranch(gitProjectID, gitBranch)

		if isSystemBranch(gitBranch) {
			isDeleteBranch = false
		} else if err != nil {
			if strings.Contains(err.Error(), "404 Branch Not Found") {
				isDeleteBranch = true
			}
		} else if len(namespace.GetAnnotations()[label_lastScaleDate]) > 0 {
			lastScaleDate, err := time.Parse(time.RFC3339, namespace.GetAnnotations()[label_lastScaleDate])
			if err != nil {
				log.Warn(err)
				logError(span, sentry.LevelWarning, nil, err, "")
			}
			if diffToNow(lastScaleDate) > *appConfig.removeBranchLastScaleDate {
				isDeleteBranch = true
			}
		} else if diffToNow(namespace.CreationTimestamp.Time) > 1 {
			isDeleteBranch = getLastCommitBranch(span, git, gitProjectID, gitBranch)
		}

		log.Debugf("gitProjectID=%s,gitBranch=%s,isDeleteBranch=%t", gitProjectID, gitBranch, isDeleteBranch)

		if isDeleteBranch {
			span.LogKV("delete branch", gitBranch)

			ch1 := make(chan httpResponse)
			q := make(url.Values)

			q.Add("namespace", ingress.Namespace)

			for k, v := range ingress.Annotations {
				if strings.HasPrefix(k, "kubernetes-manager") {
					q.Add(k[19:], v)
				}
			}

			go makeAPICall(span, "/api/deleteALL", q, ch1)

			span.LogKV("result", <-ch1)
		}
	}
}
