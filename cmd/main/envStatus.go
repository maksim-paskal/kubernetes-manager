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
	"regexp"

	sentry "github.com/getsentry/sentry-go"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getEnvStatus(w http.ResponseWriter, r *http.Request) {
	tracer := opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("getEnvStatus", ext.RPCServerOption(spanCtx))

	defer span.Finish()

	domain := r.URL.Query()["domain"]

	if len(domain) < 1 {
		http.Error(w, "domain not set", http.StatusInternalServerError)
		logError(span, sentry.LevelInfo, r, nil, "domain not set")

		return
	}

	nsRegexp := regexp.MustCompile(*appConfig.envStatusRegexp)

	res := nsRegexp.FindStringSubmatch(domain[0])

	groupIndex := -1
	nsRegexpName := nsRegexp.SubexpNames()

	for index, groupName := range nsRegexpName {
		if groupName == "namespace" {
			groupIndex = index

			break
		}
	}

	domainNamespace := ""
	if groupIndex >= 0 {
		domainNamespace = res[groupIndex]
	}

	ctx := context.Background()

	_, err := clientset.CoreV1().Namespaces().Get(ctx, domainNamespace, meta.GetOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelInfo, r, err, "")

		return
	}

	pods, err := clientset.CoreV1().Pods(domainNamespace).List(ctx, meta.ListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelInfo, r, err, "")

		return
	}

	type ResponseType struct {
		Status string
	}

	responseType := ResponseType{
		Status: "ok",
	}

	for _, pod := range pods.Items {
		ready := false

		if pod.Status.Phase == v1.PodRunning {
			for _, v := range pod.Status.Conditions {
				if v.Type == v1.PodReady && v.Status == "True" {
					ready = true
				}
			}
		}

		if !ready {
			responseType.Status = fmt.Sprintf("pod %s not ready", pod.Name)

			break
		}
	}

	w.Header().Set("Content-Type", "application/json")

	js, err := json.Marshal(responseType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelInfo, r, err, "")

		return
	}

	_, err = w.Write(js)
	//
	if err != nil {
		logError(span, sentry.LevelError, r, err, "")
	}
}
