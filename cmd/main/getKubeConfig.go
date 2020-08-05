package main

import (
	"bytes"
	"encoding/base64"
	"html/template"
	"io/ioutil"
	"net/http"

	sentry "github.com/getsentry/sentry-go"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func getKubeConfig(w http.ResponseWriter, r *http.Request) {
	var tracer = opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan("getKubeConfig", ext.RPCServerOption(spanCtx))
	defer span.Finish()

	caCRT, err := ioutil.ReadFile("/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")
		return
	}

	token, err := ioutil.ReadFile("/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logError(span, sentry.LevelError, r, err, "")
		return
	}

	var kubeConfig = `apiVersion: v1
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
		logError(span, sentry.LevelError, r, err, "")
		return
	}
	err = tmpl.Execute(&out, params)

	if err != nil {
		logError(span, sentry.LevelError, r, err, "")
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment; filename=\"kubeconfig\"")
	_, err = w.Write(out.Bytes())

	if err != nil {
		logError(span, sentry.LevelError, r, err, "")
	}
}
