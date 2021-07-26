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
	"context"
	"encoding/json"
	"net/http"

	"github.com/maksim-paskal/kubernetes-manager/pkg/api"
	logrushookopentracing "github.com/maksim-paskal/logrus-hook-opentracing"
	logrushooksentry "github.com/maksim-paskal/logrus-hook-sentry"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getPods(w http.ResponseWriter, r *http.Request) {
	tracer := opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("getPods", ext.RPCServerOption(spanCtx))

	defer span.Finish()

	namespace := r.URL.Query()["namespace"]

	if len(namespace) < 1 {
		http.Error(w, errNoNamespace.Error(), http.StatusInternalServerError)
		log.
			WithError(errNoNamespace).
			WithField(logrushookopentracing.SpanKey, span).
			WithFields(logrushooksentry.AddRequest(r)).
			Error()

		return
	}

	pods, err := api.Clientset.CoreV1().Pods(namespace[0]).List(context.TODO(), metav1.ListOptions{
		FieldSelector: "status.phase=Running",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			WithFields(logrushooksentry.AddRequest(r)).
			Error()

		return
	}

	if len(pods.Items) == 0 {
		http.Error(w, errNoPodInStatusRunning.Error(), http.StatusInternalServerError)
		log.
			WithError(errNoPodInStatusRunning).
			WithField(logrushookopentracing.SpanKey, span).
			WithFields(logrushooksentry.AddRequest(r)).
			Error()

		return
	}

	type PodContainerData struct {
		ContainerName string
	}

	type PodData struct {
		PodName       string
		PodLabels     map[string]string
		PodContainers []PodContainerData
	}

	type ResultType struct {
		Result []PodData `json:"result"`
	}

	podsData := make([]PodData, 0)

	for _, pod := range pods.Items {
		var podContainersData []PodContainerData

		for _, podContainer := range pod.Spec.Containers {
			podContainerData := PodContainerData{
				ContainerName: podContainer.Name,
			}

			podContainersData = append(podContainersData, podContainerData)
		}

		podData := PodData{
			PodName:       pod.Name,
			PodLabels:     pod.Labels,
			PodContainers: podContainersData,
		}
		podsData = append(podsData, podData)
	}

	result := ResultType{
		Result: podsData,
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
