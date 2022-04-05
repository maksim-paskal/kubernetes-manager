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
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/maksim-paskal/kubernetes-manager/pkg/api"
	logrushookopentracing "github.com/maksim-paskal/logrus-hook-opentracing"
	logrushooksentry "github.com/maksim-paskal/logrus-hook-sentry"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	log "github.com/sirupsen/logrus"
)

func deploySelectedServices(w http.ResponseWriter, r *http.Request) {
	tracer := opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("deploySelectedServices", ext.RPCServerOption(spanCtx))

	defer span.Finish()

	if err := checkParams(r, []string{"services", "namespace"}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			WithFields(logrushooksentry.AddRequest(r)).
			Error()

		return
	}

	namespace := r.URL.Query()["namespace"][0]
	services := r.URL.Query()["services"][0]

	projectPipelineDatas := strings.Split(services, ";")

	var (
		wg   sync.WaitGroup
		lock sync.Mutex
	)

	results := make([]string, 0)

	wg.Add(len(projectPipelineDatas))

	for _, projectPipelineData := range projectPipelineDatas {
		data := strings.Split(projectPipelineData, ":")

		go func(namespace string, projectID string, branch string) {
			defer wg.Done()

			var resultText string

			pipelineURL, err := api.CreateGitlabPipeline(namespace, projectID, branch)
			if err != nil {
				resultText = fmt.Sprintf("Pipeline not created %s", err.Error())
			} else {
				resultText = fmt.Sprintf("Pipeline created %s", pipelineURL)
			}

			lock.Lock()
			defer lock.Unlock()

			results = append(results, resultText)
		}(namespace, data[0], data[1])
	}

	wg.Wait()

	type ResultData struct {
		Stdout string
	}

	type ResultType struct {
		Result ResultData `json:"result"`
	}

	result := ResultType{
		Result: ResultData{
			Stdout: strings.Join(results, "\n"),
		},
	}

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
