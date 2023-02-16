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
package batch

import (
	"context"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/api"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"go.uber.org/atomic"
)

const (
	namespaceCreatedDelay    = 60 * time.Minute
	namespaceLastScaledDelay = 60 * time.Minute
)

var isStoped = *atomic.NewBool(false)

func Schedule(ctx context.Context) {
	log.Info("starting batch")

	isStoped.Store(false)

	tracer := opentracing.GlobalTracer()

	_, err := time.LoadLocation(*config.Get().BatchSheduleTimezone)
	if err != nil {
		log.WithError(err).Fatal()
	}

	for {
		<-time.After(*config.Get().BatchShedulePeriod)

		if isStoped.Load() {
			return
		}

		if ctx.Err() != nil {
			return
		}

		span := tracer.StartSpan("scheduleBatch")

		ctx, cancel := context.WithTimeout(ctx, *config.Get().BatchShedulePeriod)
		defer cancel()

		if err := Execute(ctx, span); err != nil {
			log.WithError(err).Error()
		}

		span.Finish()
	}
}

func Stop() {
	isStoped.Store(true)
}

func scaleDownALL(ctx context.Context, rootSpan opentracing.Span) error {
	tracer := opentracing.GlobalTracer()
	span := tracer.StartSpan("scaleDownALL", opentracing.ChildOf(rootSpan.Context()))

	defer span.Finish()

	if !IsScaleDownActive(time.Now()) {
		log.Debug("scaleDownALL not in period")

		return nil
	}

	environments, err := api.GetEnvironments(ctx, "")
	if err != nil {
		return errors.Wrap(err, "error listing environments")
	}

	for _, environment := range environments {
		go func(environment *api.Environment) {
			log := log.WithField("namespace", environment.Namespace)

			isScaledownDelay, err := IsScaledownDelay(time.Now(), environment)
			if err != nil {
				log.WithError(err).Error()
			} else if isScaledownDelay {
				return
			}

			log.Info("scaledown")

			err = environment.ScaleALL(ctx, 0)
			if err != nil {
				log.WithError(err).Error()
			}
		}(environment)
	}

	// scaledown servers
	servers, err := api.GetRemoteServers(ctx)
	if err != nil {
		return errors.Wrap(err, "error listing servers")
	}

	for _, server := range servers {
		go func(server *api.GetRemoteServerItem) {
			log := log.WithField("server", server.Name)
			// calculate is delay is active
			if delay, ok := server.Labels[config.LabelScaleDownDelayShort]; ok {
				scaleDelayTime, err := utils.UnixToTime(delay)
				if err != nil {
					log.WithError(err).Error()
				} else if time.Now().Before(scaleDelayTime) {
					log.Info("scale down delay is active")

					return
				}
			}

			log.Info("scaledown server")

			err := api.SetRemoteServerAction(ctx, api.SetRemoteServerActionInput{
				Cloud:  server.Cloud,
				ID:     server.ID,
				Action: api.SetRemoteServerStatusPowerOff,
			})
			if err != nil {
				log.WithError(err).Errorf("error power off server %s", server.ID)
			}
		}(server)
	}

	return nil
}

func Execute(ctx context.Context, rootSpan opentracing.Span) error {
	tracer := opentracing.GlobalTracer()
	span := tracer.StartSpan("batch", opentracing.ChildOf(rootSpan.Context()))

	defer span.Finish()

	if err := scaleDownALL(ctx, span); err != nil {
		log.WithError(err).Error()
	}

	environments, err := api.GetEnvironments(ctx, "")
	if err != nil {
		return errors.Wrap(err, "error list ingress")
	}

	for _, environment := range environments {
		log := log.WithFields(log.Fields{
			"namespace": environment.Namespace,
		})

		if environment.IsSystemNamespace() {
			log.Debugf("%s is system namespace", environment.Namespace)

			continue
		}

		// delete temporary tokens in namespace
		if err := environment.DeleteTemporaryTokens(ctx); err != nil {
			log.WithError(err).Error()
		}

		reason, description := environment.IsStaled()

		log.WithField("reason", reason).Debug(description)

		if reason != api.StaledReasonNone {
			deleteALLResult := environment.DeleteALL(ctx)

			log.Info(deleteALLResult.JSON())
		}
	}

	return nil
}
