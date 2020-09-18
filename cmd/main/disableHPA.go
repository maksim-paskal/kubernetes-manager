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

	sentry "github.com/getsentry/sentry-go"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func disableHPA(w http.ResponseWriter, r *http.Request) {
	tracer := opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("disableHPA", ext.RPCServerOption(spanCtx))

	defer span.Finish()

	namespace := r.URL.Query()["namespace"]

	if len(namespace) < 1 {
		http.Error(w, MessageNamespaceNotSet, http.StatusInternalServerError)
		logError(span, sentry.LevelInfo, r, nil, MessageNamespaceNotSet)

		return
	}

	if isSystemNamespace(namespace[0]) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte("{status:'ok',warning:'namespace can not disable autoscale'}"))
		if err != nil { //nolint:wsl
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logError(span, sentry.LevelError, r, err, "")
		}

		return
	}

	hpas, err := clientset.AutoscalingV1().HorizontalPodAutoscalers(namespace[0]).List(metav1.ListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")

		return
	}

	GracePeriodSeconds := int64(0)

	opt := &metav1.DeleteOptions{
		GracePeriodSeconds: &GracePeriodSeconds,
	}

	for _, hpa := range hpas.Items {
		err := clientset.AutoscalingV1().HorizontalPodAutoscalers(namespace[0]).Delete(hpa.Name, opt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logError(span, sentry.LevelError, r, err, "")

			return
		}
	}

	ch1 := make(chan httpResponse)
	q := make(url.Values)

	q.Add("namespace", namespace[0])
	q.Add("version", "1")
	q.Add("replicas", "1")

	go makeAPICall(span, "/api/scaleNamespace", q, ch1)

	type ResultData struct {
		Stdout string
		Result httpResponse
	}

	type ResultType struct {
		ScaleNamespaceResult ResultData `json:"result"`
	}

	result := ResultType{
		ScaleNamespaceResult: ResultData{
			Result: (<-ch1),
			Stdout: "Autoscale disabled",
		},
	}

	js, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")

		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")

		return
	}
}
