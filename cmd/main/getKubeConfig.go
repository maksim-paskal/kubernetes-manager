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
	"bytes"
	"encoding/base64"
	"html/template"
	"io/ioutil"
	"net/http"

	logrushookopentracing "github.com/maksim-paskal/logrus-hook-opentracing"
	logrushooksentry "github.com/maksim-paskal/logrus-hook-sentry"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	log "github.com/sirupsen/logrus"
)

func getKubeConfig(w http.ResponseWriter, r *http.Request) {
	tracer := opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("getKubeConfig", ext.RPCServerOption(spanCtx))

	defer span.Finish()

	caCRT, err := ioutil.ReadFile("/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			WithField(logrushooksentry.RequestKey, r).
			Error()

		return
	}

	token, err := ioutil.ReadFile("/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			WithField(logrushooksentry.RequestKey, r).
			Error()

		return
	}

	kubeConfig := `apiVersion: v1
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: {{ .ClusterServer }}
  name: kubernetes-manager
contexts:
- context:
    cluster: kubernetes-manager
    user: kubernetes-manager
  name: kubernetes-manager
current-context: kubernetes-manager
kind: Config
preferences: {}
users:
- name: kubernetes-manager
  user:
    token: {{ .UserToken }}`

	type Inventory struct {
		ClusterCAD    string
		ClusterServer string
		UserToken     string
	}

	params := Inventory{
		ClusterCAD:    base64.StdEncoding.EncodeToString(caCRT),
		ClusterServer: *appConfig.kubeconfigServer,
		UserToken:     string(token),
	}

	var out bytes.Buffer

	tmpl, err := template.New("kubeconfig").Parse(kubeConfig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			WithField(logrushooksentry.RequestKey, r).
			Error()

		return
	}

	err = tmpl.Execute(&out, params)

	if err != nil {
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			WithField(logrushooksentry.RequestKey, r).
			Error()
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment; filename=\"kubeconfig\"")
	_, err = w.Write(out.Bytes())

	if err != nil {
		log.
			WithError(err).
			WithField(logrushookopentracing.SpanKey, span).
			WithField(logrushooksentry.RequestKey, r).
			Error()
	}
}
