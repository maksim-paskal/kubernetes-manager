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
package web

import (
	"encoding/json"
	"net/http"

	"github.com/maksim-paskal/kubernetes-manager/pkg/api"
	logrushookopentracing "github.com/maksim-paskal/logrus-hook-opentracing"
	logrushooksentry "github.com/maksim-paskal/logrus-hook-sentry"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	log "github.com/sirupsen/logrus"
)

func execCommands(w http.ResponseWriter, r *http.Request) {
	tracer := opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("execCommands", ext.RPCServerOption(spanCtx))

	defer span.Finish()

	if err := checkParams(r, []string{"cmd"}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			WithFields(logrushooksentry.AddRequest(r)).
			Error()

		return
	}

	cmd := r.URL.Query()["cmd"]

	_, ok := getInfoDBCommands[cmd[0]]
	if !ok {
		http.Error(w, errNoComandFound.Error(), http.StatusInternalServerError)
		log.
			WithError(errNoComandFound).
			WithField(logrushookopentracing.SpanKey, span).
			WithFields(logrushooksentry.AddRequest(r)).
			Error()

		return
	}

	podExecute := getInfoDBCommands[cmd[0]]

	if podExecute.beforeExecute != nil {
		err := podExecute.beforeExecute(&podExecute.param, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.
				WithError(err).
				WithField(logrushookopentracing.SpanKey, span).
				WithFields(logrushooksentry.AddRequest(r)).
				Error()

			return
		}
	}

	execResults, err := execContainer(span, podExecute.param)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			WithFields(logrushooksentry.AddRequest(r)).
			Error()

		return
	}

	type ResultType struct {
		Result *api.ExecContainerResults `json:"result"`
	}

	if podExecute.filterStdout != nil {
		execResults.Stdout = podExecute.filterStdout(podExecute.param, execResults.Stdout)
	}

	result := ResultType{
		Result: execResults,
	}

	span.LogKV("result", result)

	js, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			WithFields(logrushooksentry.AddRequest(r)).
			Error()

		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)

	if err != nil {
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			WithFields(logrushooksentry.AddRequest(r)).
			Error()
	}
}
