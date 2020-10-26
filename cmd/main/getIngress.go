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
	"net/http"
	"time"

	sentry "github.com/getsentry/sentry-go"
	utils "github.com/maksim-paskal/utils-go"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/common/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getIngress(w http.ResponseWriter, r *http.Request) {
	tracer := opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("getIngress", ext.RPCServerOption(spanCtx))

	defer span.Finish()

	type IngressList struct {
		Namespace               string
		NamespaceStatus         string
		NamespaceCreated        string
		NamespaceCreatedDays    int
		NamespaceLastScaled     string
		NamespaceLastScaledDays int
		IngressName             string
		IngressAnotations       map[string]string
		IngressLabels           map[string]string
		Hosts                   []string
		GitBranch               string
		RunningPodsCount        int
	}

	type IngressListResult struct {
		Result []IngressList `json:"result"`
	}

	var result IngressListResult

	opt := metav1.ListOptions{
		LabelSelector: *appConfig.ingressFilter,
	}
	if *appConfig.ingressNoFiltration {
		opt = metav1.ListOptions{}
	}

	span.LogKV("event", "start ingress list")

	ingresss, err := clientset.ExtensionsV1beta1().Ingresses("").List(context.TODO(), opt)

	span.LogKV("event", "end ingress list")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")

		return
	}

	span.LogKV("event", "start range")

	for _, ingress := range ingresss.Items {
		var item IngressList

		span.LogKV("event", "search namespace="+ingress.Namespace)
		namespace, err := clientset.CoreV1().Namespaces().Get(context.TODO(), ingress.Namespace, metav1.GetOptions{})
		if err != nil { //nolint:wsl
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logError(span, sentry.LevelError, r, err, "")

			return
		}

		item.GitBranch = ingress.Annotations[labelGitBranch]

		if len(namespace.GetAnnotations()[labelLastScaleDate]) > 0 {
			lastScaleDate, err := time.Parse(time.RFC3339, namespace.GetAnnotations()[labelLastScaleDate])
			if err != nil {
				log.Warn(err)
				logError(span, sentry.LevelWarning, r, err, "")
			} else {
				item.NamespaceLastScaled = lastScaleDate.String()
				item.NamespaceLastScaledDays = diffToNow(lastScaleDate)
			}
		}

		item.Namespace = namespace.Name
		item.NamespaceStatus = string(namespace.Status.Phase)
		item.NamespaceCreated = namespace.CreationTimestamp.String()
		item.RunningPodsCount = -1
		item.NamespaceCreatedDays = diffToNow(namespace.CreationTimestamp.Time)

		item.IngressName = ingress.Name
		item.IngressAnotations = ingress.Annotations
		item.IngressLabels = ingress.Labels

		for _, rule := range ingress.Spec.Rules {
			host := "http://" + rule.Host
			if !utils.StringInSlice(host, item.Hosts) {
				item.Hosts = append(item.Hosts, host)
			}
		}

		if len(item.Hosts) > 0 {
			result.Result = append(result.Result, item)
		}
	}

	span.LogKV("event", "end range")
	span.LogKV("event", "result", result)
	js, err := json.Marshal(result)
	if err != nil { //nolint:wsl
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "max-age=10")
	_, err = w.Write(js)

	if err != nil {
		logError(span, sentry.LevelError, r, err, "")
	}
}
