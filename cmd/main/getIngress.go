package main

import (
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
	var tracer = opentracing.GlobalTracer()
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
	ingresss, err := clientset.ExtensionsV1beta1().Ingresses("").List(opt)
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
		namespace, err := clientset.CoreV1().Namespaces().Get(ingress.Namespace, metav1.GetOptions{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logError(span, sentry.LevelError, r, err, "")
			return
		}

		item.GitBranch = ingress.Annotations[label_gitBranch]

		if len(namespace.GetAnnotations()[label_lastScaleDate]) > 0 {
			lastScaleDate, err := time.Parse(time.RFC3339, namespace.GetAnnotations()[label_lastScaleDate])
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
	if err != nil {
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
