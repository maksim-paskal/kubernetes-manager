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
	"os/signal"
	"syscall"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/api"
	"github.com/maksim-paskal/kubernetes-manager/pkg/batch"
	"github.com/maksim-paskal/kubernetes-manager/pkg/cache"
	"github.com/maksim-paskal/kubernetes-manager/pkg/client"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/maksim-paskal/kubernetes-manager/pkg/web"
	"github.com/maksim-paskal/kubernetes-manager/pkg/webhook"
	logrushooksentry "github.com/maksim-paskal/logrus-hook-sentry"
	log "github.com/sirupsen/logrus"
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

	// get background context
	ctx, cancel := context.WithCancel(context.Background())
	signalChanInterrupt := make(chan os.Signal, 1)
	signal.Notify(signalChanInterrupt, syscall.SIGINT, syscall.SIGTERM)

	var err error

	if err = config.Load(); err != nil {
		log.WithError(err).Fatal()
	}

	logLevel, err := log.ParseLevel(*config.Get().LogLevel)
	if err != nil {
		log.WithError(err).Fatal()
	}

	log.SetLevel(logLevel)
	log.SetReportCaller(true)

	if err = config.CheckConfig(); err != nil {
		log.WithError(err).Fatal()
	}

	if err = webhook.CheckConfig(); err != nil {
		log.WithError(err).Fatal()
	}

	log.Debugf("Using config:\n%s", config.Get().String())

	if err := cache.Init(ctx, cache.ProviderName(config.Get().Cache.Type), config.Get().Cache.Config); err != nil {
		log.WithError(err).Fatal()
	}

	hookSentry, err := logrushooksentry.NewHook(ctx, logrushooksentry.Options{
		Release: config.GetVersion(),
	})
	if err != nil {
		log.WithError(err).Fatal()
	}

	log.AddHook(hookSentry)

	if err := telemetry.Init(ctx); err != nil {
		log.WithError(err).Fatal()
	}

	log.Infof("Starting %s %s...", config.Namespace, config.GetVersion())

	err = client.Init()
	if err != nil {
		log.WithError(err).Fatal()
	}

	if *config.Get().BatchEnabled {
		go RunLeaderElection(ctx)
	}

	go func() {
		select {
		case <-signalChanInterrupt:
			log.Warn("Got interruption signal...")
			cancel()

			<-signalChanInterrupt
			log.Error("Got second interruption signal...")
			os.Exit(1)
		case <-ctx.Done():
		}
	}()

	log.RegisterExitHandler(func() {
		cancel()
		time.Sleep(config.Get().GetGracefulShutdown())
	})

	go web.StartServer(ctx)

	<-ctx.Done()

	log.Info("Shutting down...")

	time.Sleep(config.Get().GetGracefulShutdown())
}

func RunLeaderElection(ctx context.Context) {
	ctx, span := telemetry.Start(ctx, "api.RunLeaderElection")
	defer span.End()

	lock, err := api.GetLeaseLock(*config.Get().PodNamespace, *config.Get().PodName)
	if err != nil {
		log.WithError(err).Fatal()

		return
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   defaultLeaseDuration,
		RenewDeadline:   defaultRenewDeadline,
		RetryPeriod:     defaultRetryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				ctx, span := telemetry.Start(ctx, "api.OnStartedLeading")
				defer span.End()

				batch.Schedule(ctx)
			},
			OnStoppedLeading: func() {
				cancel()
			},
		},
	})
}
