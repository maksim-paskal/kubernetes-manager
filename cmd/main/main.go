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
	"fmt"
	"net/http"

	//nolint:gosec
	_ "net/http/pprof"

	logrushookopentracing "github.com/maksim-paskal/logrus-hook-opentracing"
	logrushooksentry "github.com/maksim-paskal/logrus-hook-sentry"
	opentracing "github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics"
	"gopkg.in/alecthomas/kingpin.v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	gitVersion = "dev"
	clientset  *kubernetes.Clientset
	restconfig *rest.Config
)

func main() {
	kingpin.Version(appConfig.Version)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	var err error

	logLevel, err := log.ParseLevel(*appConfig.logLevel)
	if err != nil {
		log.WithError(err).Fatal()
	}

	log.SetLevel(logLevel)

	if logLevel >= log.DebugLevel {
		log.SetReportCaller(true)
	}

	hookSentry, err := logrushooksentry.NewHook(logrushooksentry.Options{
		Release: appConfig.Version,
	})
	if err != nil {
		log.WithError(err).Fatal()
	}

	log.AddHook(hookSentry)
	defer hookSentry.Stop()

	hookTracing, err := logrushookopentracing.NewHook(logrushookopentracing.Options{})
	if err != nil {
		log.WithError(err).Fatal()
	}

	log.AddHook(hookTracing)

	log.Infof("Starting kubernetes-manager %s...", appConfig.Version)

	if len(*appConfig.kubeconfigPath) > 0 {
		restconfig, err = clientcmd.BuildConfigFromFlags("", *appConfig.kubeconfigPath)
		if err != nil {
			log.WithError(err).Fatal()
		}
	} else {
		log.Info("No kubeconfig file use incluster")
		restconfig, err = rest.InClusterConfig()
		if err != nil {
			log.WithError(err).Fatal()
		}
	}

	clientset, err = kubernetes.NewForConfig(restconfig)
	if err != nil {
		log.WithError(err).Fatal()
	}

	cfg, err := jaegercfg.FromEnv()
	if err != nil {
		log.WithError(err).Fatal("Could not parse Jaeger env vars")
	}

	cfg.ServiceName = "kubernetes-manager"
	cfg.Sampler.Type = jaeger.SamplerTypeConst
	cfg.Sampler.Param = 1
	cfg.Reporter.LogSpans = true

	jLogger := LogrusAdapter{}
	jMetricsFactory := metrics.NullFactory

	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)

	opentracing.SetGlobalTracer(tracer)

	if err != nil {
		log.WithError(err).Fatal("Could not initialize jaeger tracer")
	}
	defer closer.Close()

	if *appConfig.mode == "batch" {
		span := tracer.StartSpan("main")

		defer span.Finish()

		batch(span)

		return
	}

	if *appConfig.mode == "cleanOldTags" {
		span := tracer.StartSpan("main")

		defer span.Finish()

		cleanOldTags(span)

		return
	}

	go scheduleBatch()

	log.Info(fmt.Sprintf("Starting on port %d...", *appConfig.port))
	fs := http.FileServer(http.Dir(*appConfig.frontDist))

	http.Handle("/", fs)
	http.HandleFunc("/_nuxt/", serveFiles)
	http.HandleFunc("/api/getIngress", getIngress)
	http.HandleFunc("/api/getNamespace", getNamespace)
	http.HandleFunc("/api/deleteNamespace", deleteNamespace)
	http.HandleFunc("/api/deleteRegistryTag", deleteRegistryTag)
	http.HandleFunc("/api/deletePod", deletePod)
	http.HandleFunc("/api/exec", execCommands)
	http.HandleFunc("/api/deleteALL", deleteALL)
	http.HandleFunc("/api/executeBatch", executeBatch)
	http.HandleFunc("/getKubeConfig", getKubeConfig)
	http.HandleFunc("/api/scaleNamespace", scaleNamespace)
	http.HandleFunc("/api/getRunningPodsCount", getRunningPodsCount)
	http.HandleFunc("/api/version", getAPIversion)
	http.HandleFunc("/api/getPods", getPods)
	http.HandleFunc("/api/debug", getDebug)
	http.HandleFunc("/api/disableHPA", disableHPA)

	err = http.ListenAndServe(fmt.Sprintf(":%d", *appConfig.port), nil)
	if err != nil {
		log.WithError(err).Fatal()
	}
}
