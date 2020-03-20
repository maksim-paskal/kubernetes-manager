/*
Copyright paskal.maksim@gmail.com
Licensed under the Apache License, Version 2.0 (the "License");
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
	"net/url"
	"strings"

	sentry "github.com/getsentry/sentry-go"
	opentracing "github.com/opentracing/opentracing-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func pauseALL(rootSpan opentracing.Span) {
	var tracer = opentracing.GlobalTracer()
	span := tracer.StartSpan("pauseALL", opentracing.ChildOf(rootSpan.Context()))
	defer span.Finish()

	opt := metav1.ListOptions{
		LabelSelector: *appConfig.ingressFilter,
	}

	span.LogKV("event", "search namespace")
	ingresss, err := clientset.ExtensionsV1beta1().Ingresses("").List(opt)
	span.LogKV("event", "search complete")

	if err != nil {
		logError(span, sentry.LevelError, nil, err, "")
		return
	}

	for _, ingress := range ingresss.Items {
		span1 := tracer.StartSpan("pause-namespace", opentracing.ChildOf(span.Context()))
		defer span1.Finish()
		span1.LogKV("namespace", ingress.Namespace)

		ch1 := make(chan httpResponse)
		q := make(url.Values)

		q.Add("namespace", ingress.Namespace)

		for k, v := range ingress.Annotations {
			if strings.HasPrefix(k, "kubernetes-manager") {
				q.Add(k[19:], v)
			}
		}

		go makeAPICall(span1, "/api/scaleNamespace", q, ch1)

		span1.LogKV("result", <-ch1)
	}
}
