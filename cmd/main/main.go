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
	"os"
	"time"

	sentry "github.com/getsentry/sentry-go"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
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
	gitVersion string = "dev"
	buildTime  string
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
		log.Panic(err)
	}

	log.SetLevel(logLevel)

	if logLevel >= log.DebugLevel {
		log.SetReportCaller(true)
	}

	log.Infof("Starting kubernetes-manager %s", appConfig.Version)

	if len(os.Getenv("SENTRY_DSN")) > 0 {
		log.Debug("Use Sentry logging...")

		err = sentry.Init(sentry.ClientOptions{
			Release: fmt.Sprintf("%s-%s", appConfig.Version, buildTime),
		})

		if err != nil {
			fmt.Printf("Sentry initialization failed: %v\n", err)
		}
	}

	if len(*appConfig.kubeconfigPath) > 0 {
		restconfig, err = clientcmd.BuildConfigFromFlags("", *appConfig.kubeconfigPath)
		if err != nil {
			sentry.CaptureException(err)
			sentry.Flush(time.Second)

			log.Panic(err.Error())
		}
	} else {
		log.Info("No kubeconfig file use incluster")
		restconfig, err = rest.InClusterConfig()
		if err != nil {
			log.Panic(err.Error())
		}
	}

	clientset, err = kubernetes.NewForConfig(restconfig)
	if err != nil {
		sentry.CaptureException(err)
		sentry.Flush(time.Second)

		log.Panic(err.Error())
	}

	cfg, err := jaegercfg.FromEnv()
	if err != nil {
		sentry.CaptureException(err)
		sentry.Flush(time.Second)

		log.Panicf("Could not parse Jaeger env vars: %s", err.Error())
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
		sentry.CaptureException(err)
		sentry.Flush(time.Second)

		log.Panicf("Could not initialize jaeger tracer: %s", err.Error())
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
		sentry.CaptureException(err)
		sentry.Flush(time.Second)

		log.Fatal(errors.Wrap(err, "http.ListenAndServe")) //nolint:gocritic
	}
}
