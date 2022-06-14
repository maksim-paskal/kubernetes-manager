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
	"flag"
	"fmt"
	_ "net/http/pprof" //nolint:gosec
	"os"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/api"
	"github.com/maksim-paskal/kubernetes-manager/pkg/batch"
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
	"k8s.io/client-go/tools/leaderelection"
)

var version = flag.Bool("version", false, "version")

const (
	defaultLeaseDuration = 15 * time.Second
	defaultRenewDeadline = 10 * time.Second
	defaultRetryPeriod   = 2 * time.Second
)

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

	log.Infof("Starting %s %s...", config.Namespace, config.GetVersion())

	err = api.Init()
	if err != nil {
		log.WithError(err).Fatal()
	}

	cfg, err := jaegercfg.FromEnv()
	if err != nil {
		log.WithError(err).Fatal("Could not parse Jaeger env vars")
	}

	cfg.ServiceName = config.Namespace
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

	go RunLeaderElection()

	web.StartServer()
}

func RunLeaderElection() {
	lock, err := api.GetLeaseLock(*config.Get().PodNamespace, *config.Get().PodName)
	if err != nil {
		log.WithError(err).Fatal()

		return
	}

	leaderelection.RunOrDie(context.Background(), leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   defaultLeaseDuration,
		RenewDeadline:   defaultRenewDeadline,
		RetryPeriod:     defaultRetryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(c context.Context) {
				batch.Schedule()
			},
			OnStoppedLeading: func() {
				batch.Stop()
			},
		},
	})
}
