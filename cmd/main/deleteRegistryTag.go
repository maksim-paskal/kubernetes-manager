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
package main

import (
	"fmt"
	"net/http"

	logrushookopentracing "github.com/maksim-paskal/logrus-hook-opentracing"
	logrushooksentry "github.com/maksim-paskal/logrus-hook-sentry"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

func deleteRegistryTag(w http.ResponseWriter, r *http.Request) {
	tracer := opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("deleteRegistryTag", ext.RPCServerOption(spanCtx))

	defer span.Finish()

	tag := r.URL.Query()["tag"]

	if len(tag) < 1 {
		http.Error(w, ErrNoTag.Error(), http.StatusInternalServerError)
		log.
			WithError(ErrNoTag).
			WithField(logrushookopentracing.SpanKey, span).
			WithField(logrushooksentry.RequestKey, r).
			Error()

		return
	}

	if isSystemBranch(tag[0]) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte("{status:'ok',warning:'registry tag can not be deleted'}"))
		if err != nil { //nolint:wsl
			log.
				WithError(err).
				WithField(logrushookopentracing.SpanKey, span).
				WithField(logrushooksentry.RequestKey, r).
				Error()
		}

		return
	}

	projectID := r.URL.Query()["projectID"]

	if len(projectID) < 1 {
		http.Error(w, ErrNoProjectID.Error(), http.StatusInternalServerError)
		log.
			WithError(ErrNoProjectID).
			WithField(logrushookopentracing.SpanKey, span).
			WithField(logrushooksentry.RequestKey, r).
			Error()

		return
	}

	span.LogKV("params", fmt.Sprintf("projectID=%s,tag=%s", projectID[0], tag[0]))

	git, err := gitlab.NewClient(*appConfig.gitlabToken, gitlab.WithBaseURL(*appConfig.gitlabURL))
	if err != nil {
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			WithField(logrushooksentry.RequestKey, r).
			Error()
	}

	span.LogKV("event", "ListRegistryRepositories")

	gitRepos, _, err := git.ContainerRegistry.ListRegistryRepositories(projectID[0], nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			WithField(logrushooksentry.RequestKey, r).
			Error()

		return
	}

	for _, gitRepo := range gitRepos {
		span.LogKV("DeleteRegistryRepositoryTag", fmt.Sprintf("gitRepo.ID=%d", gitRepo.ID))

		_, err := git.ContainerRegistry.DeleteRegistryRepositoryTag(projectID[0], gitRepo.ID, tag[0])
		if err != nil {
			span.LogKV("warning", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write([]byte("{status:'ok'}"))

	if err != nil {
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			WithField(logrushooksentry.RequestKey, r).
			Error()
	}
}
