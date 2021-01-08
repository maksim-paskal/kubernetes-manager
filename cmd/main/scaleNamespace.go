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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	logrushookopentracing "github.com/maksim-paskal/logrus-hook-opentracing"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func scaleNamespace(w http.ResponseWriter, r *http.Request) {
	tracer := opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("scaleNamespace", ext.RPCServerOption(spanCtx))

	defer span.Finish()

	namespace := r.URL.Query()["namespace"]
	version := 0

	if len(namespace) < 1 {
		http.Error(w, ErrNoNamespace.Error(), http.StatusInternalServerError)
		log.
			WithError(ErrNoNamespace).
			WithField(logrushookopentracing.SpanKey, span).
			Error()

		return
	}

	if len(r.URL.Query()["version"]) == 1 {
		var err error

		version, err = strconv.Atoi(r.URL.Query()["version"][0])
		if err != nil {
			log.
				WithError(err).
				WithField(logrushookopentracing.SpanKey, span).
				Warn("can not parse version ", version)
		}
	}

	replicas := r.URL.Query()["replicas"]

	if len(replicas) < 1 {
		http.Error(w, ErrNoReplicas.Error(), http.StatusInternalServerError)
		log.
			WithError(ErrNoReplicas).
			WithField(logrushookopentracing.SpanKey, span).
			Error()

		return
	}

	ds, err := clientset.AppsV1().Deployments(namespace[0]).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			Error()

		return
	}
	//nolint:dupl
	for _, d := range ds.Items {
		dps, err := clientset.AppsV1().Deployments(namespace[0]).Get(context.TODO(), d.Name, metav1.GetOptions{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.
				WithError(err).
				WithField(logrushookopentracing.SpanKey, span).
				Error()

			return
		}

		i, err := strconv.ParseInt(replicas[0], 10, 32)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.
				WithError(err).
				WithField(logrushookopentracing.SpanKey, span).
				Error()

			return
		}

		i32 := int32(i)
		dps.Spec.Replicas = &i32
		_, errUpdate := clientset.AppsV1().Deployments(namespace[0]).Update(context.TODO(), dps, metav1.UpdateOptions{})

		if errUpdate != nil {
			http.Error(w, errUpdate.Error(), http.StatusInternalServerError)
			log.
				WithError(errUpdate).
				WithField(logrushookopentracing.SpanKey, span).
				Error()

			return
		}
	}

	// scale statefullsets
	if version > 0 {
		sf, err := clientset.AppsV1().StatefulSets(namespace[0]).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.
				WithError(err).
				WithField(logrushookopentracing.SpanKey, span).
				Error()

			return
		}

		//nolint:dupl
		for _, s := range sf.Items {
			ss, err := clientset.AppsV1().StatefulSets(namespace[0]).Get(context.TODO(), s.Name, metav1.GetOptions{})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.
					WithError(err).
					WithField(logrushookopentracing.SpanKey, span).
					Error()

				return
			}

			i, err := strconv.ParseInt(replicas[0], 10, 32)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.
					WithError(err).
					WithField(logrushookopentracing.SpanKey, span).
					Error()

				return
			}

			i32 := int32(i)
			ss.Spec.Replicas = &i32
			_, errUpdate := clientset.AppsV1().StatefulSets(namespace[0]).Update(context.TODO(), ss, metav1.UpdateOptions{})

			if errUpdate != nil {
				http.Error(w, errUpdate.Error(), http.StatusInternalServerError)
				log.
					WithError(errUpdate).
					WithField(logrushookopentracing.SpanKey, span).
					Error()

				return
			}
		}
	}

	type ResultData struct {
		Stdout string
	}

	type ResultType struct {
		Result ResultData `json:"result"`
	}

	result := ResultType{
		Result: ResultData{
			Stdout: fmt.Sprintf("all Deployments/StatefulSets in namespace scaled to %s", replicas[0]),
		},
	}
	js, err := json.Marshal(result)
	if err != nil { //nolint:wsl
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			Error()

		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)

	if err != nil {
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			Error()
	}

	/*patch*/
	type metadataStringValue struct {
		Annotations map[string]string `json:"annotations"`
	}

	type patchStringValue struct {
		Metadata metadataStringValue `json:"metadata"`
	}

	payload := patchStringValue{
		Metadata: metadataStringValue{
			Annotations: map[string]string{labelLastScaleDate: time.Now().Format(time.RFC3339)},
		},
	}
	payloadBytes, _ := json.Marshal(payload)
	ns := clientset.CoreV1().Namespaces()
	_, err = ns.Patch(context.TODO(), namespace[0], types.StrategicMergePatchType, payloadBytes, metav1.PatchOptions{})

	if err != nil {
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			Warn()
	}
}
