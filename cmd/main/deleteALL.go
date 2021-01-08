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
	"encoding/json"
	"net/http"
	"net/url"

	logrushookopentracing "github.com/maksim-paskal/logrus-hook-opentracing"
	logrushooksentry "github.com/maksim-paskal/logrus-hook-sentry"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	log "github.com/sirupsen/logrus"
)

func deleteALL(w http.ResponseWriter, r *http.Request) {
	tracer := opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("deleteALL", ext.RPCServerOption(spanCtx))

	defer span.Finish()

	namespace := r.URL.Query()["namespace"]

	if len(namespace) != 1 {
		http.Error(w, ErrNoNamespace.Error(), http.StatusInternalServerError)
		log.
			WithError(ErrNoNamespace).
			WithField(logrushookopentracing.SpanKey, span).
			WithField(logrushooksentry.RequestKey, r).
			Error()

		return
	}

	if isSystemNamespace(namespace[0]) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte("{status:'ok',warning:'namespace can not be deleted'}"))
		if err != nil { //nolint:wsl
			log.
				WithError(err).
				WithField(logrushookopentracing.SpanKey, span).
				WithField(logrushooksentry.RequestKey, r).
				Error()
		}

		return
	}

	log.
		WithError(ErrUserDeleteALL).
		WithField(logrushookopentracing.SpanKey, span).
		WithField(logrushooksentry.RequestKey, r).
		Warn()

	type ResultData struct {
		DeleteNamespaceResultBody   httpResponse
		DeleteRegistryTagResultBody httpResponse
	}

	type ResultType struct {
		Result ResultData `json:"result"`
	}

	result := ResultType{
		Result: ResultData{},
	}

	ch3 := make(chan httpResponse)
	q := make(url.Values)

	q.Add("namespace", namespace[0])

	go makeAPICall(span, "/api/deleteNamespace", q, ch3)

	result.Result.DeleteNamespaceResultBody = (<-ch3)

	projectID := r.URL.Query()["git-project-id"]
	tag := r.URL.Query()["registry-tag"]

	if len(projectID) == 1 && len(tag) == 1 {
		ch4 := make(chan httpResponse)
		q = make(url.Values)
		q.Add("projectID", r.URL.Query()["git-project-id"][0])
		q.Add("tag", r.URL.Query()["registry-tag"][0])

		go makeAPICall(span, "/api/deleteRegistryTag", q, ch4)

		result.Result.DeleteRegistryTagResultBody = (<-ch4)
	} else {
		result.Result.DeleteRegistryTagResultBody = httpResponse{
			Status: "not executed",
			Body:   "projectID or tag not set",
		}
	}

	span.LogKV("result", result)
	js, err := json.Marshal(result)
	if err != nil { //nolint:wsl
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			WithField(logrushooksentry.RequestKey, r).
			Error()

		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)

	if err != nil {
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			WithField(logrushooksentry.RequestKey, r).
			Error()
	}
}
