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
	"fmt"
	"net/http"
	"strings"

	"github.com/maksim-paskal/kubernetes-manager/pkg/api"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	logrushookopentracing "github.com/maksim-paskal/logrus-hook-opentracing"
	logrushooksentry "github.com/maksim-paskal/logrus-hook-sentry"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deletePod(w http.ResponseWriter, r *http.Request) {
	tracer := opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("deletePod", ext.RPCServerOption(spanCtx))

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

	GracePeriodSeconds := int64(0)

	opt := &metav1.DeleteOptions{
		GracePeriodSeconds: &GracePeriodSeconds,
	}

	var podName string

	LabelSelector := r.URL.Query()["LabelSelector"]
	pod := r.URL.Query()["pod"]

	if len(pod) > 0 {
		podinfo := strings.Split(pod[0], ":")

		if len(podinfo) != config.KeyValueLength {
			http.Error(w, errNoPodSelected.Error(), http.StatusInternalServerError)
			log.
				WithError(errNoPodSelected).
				WithField(logrushookopentracing.SpanKey, span).
				WithFields(logrushooksentry.AddRequest(r)).
				Error()

			return
		}

		podName = podinfo[0]
	} else {
		if len(LabelSelector) < 1 {
			http.Error(w, errNoLabelSelector.Error(), http.StatusInternalServerError)
			log.
				WithError(errNoLabelSelector).
				WithField(logrushookopentracing.SpanKey, span).
				WithFields(logrushooksentry.AddRequest(r)).
				Error()

			return
		}

		pods, err1 := api.Clientset.CoreV1().Pods(namespace[0]).List(context.TODO(), metav1.ListOptions{
			LabelSelector: LabelSelector[0],
			FieldSelector: "status.phase=Running",
		})

		if err1 != nil {
			http.Error(w, err1.Error(), http.StatusInternalServerError)
			log.
				WithError(err1).
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

		podName = pods.Items[0].Name
	}

	err2 := api.Clientset.CoreV1().Pods(namespace[0]).Delete(context.TODO(), podName, *opt)

	if err2 != nil {
		http.Error(w, err2.Error(), http.StatusInternalServerError)
		log.
			WithError(err2).
			WithField(logrushookopentracing.SpanKey, span).
			WithFields(logrushooksentry.AddRequest(r)).
			Error()

		return
	}

	type ResultData struct {
		Stdout string
	}

	type ResultType struct {
		Result ResultData `json:"result"`
	}

	result := ResultType{
		Result: ResultData{
			Stdout: fmt.Sprintf("deleted %s pod", podName),
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
