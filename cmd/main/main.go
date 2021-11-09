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
	"flag"
	"fmt"

	//nolint:gosec
	_ "net/http/pprof"
	"os"

	"github.com/maksim-paskal/kubernetes-manager/pkg/api"
	"github.com/maksim-paskal/kubernetes-manager/pkg/batch"
	"github.com/maksim-paskal/kubernetes-manager/pkg/cleanoldtags"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/maksim-paskal/kubernetes-manager/pkg/web"
	logrushookopentracing "github.com/maksim-paskal/logrus-hook-opentracing"
	logrushooksentry "github.com/maksim-paskal/logrus-hook-sentry"
	opentracing "github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics"
)

var version = flag.Bool("version", false, "version")

func main() {
	flag.Parse()

	if *version {
		fmt.Println(config.GetVersion()) //nolint:forbidigo
		os.Exit(0)
	}

	var err error

	if err = config.Load(); err != nil {
		log.WithError(err).Fatal()
	}

	if err = config.CheckConfig(); err != nil {
		log.WithError(err).Fatal()
	}

	logLevel, err := log.ParseLevel(*config.Get().LogLevel)
	if err != nil {
		log.WithError(err).Fatal()
	}

	log.SetLevel(logLevel)
	log.SetReportCaller(true)

	log.Debugf("Using config:\n%s", config.String())

	hookSentry, err := logrushooksentry.NewHook(logrushooksentry.Options{
		Release: config.GetVersion(),
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

	log.Infof("Starting kubernetes-manager %s...", config.GetVersion())

	err = api.Init()
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

	jLogger := utils.JaegerLogs{}
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

	if *config.Get().ExecuteCleanOldTags {
		span := tracer.StartSpan("main")

		defer span.Finish()

		if err := cleanoldtags.Execute(span); err != nil {
			log.WithError(err).Error()
		}

		return
	}

	if *config.Get().ExecuteBatch {
		span := tracer.StartSpan("main")

		defer span.Finish()

		if err := batch.Execute(span); err != nil {
			log.WithError(err).Error()
		}

		return
	}

	if *config.Get().BatchShedule {
		go batch.Schedule()
	}

	web.StartServer()
}
