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
	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/maksim-paskal/kubernetes-manager/pkg/webhook"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

const maxScaleDownDuration = 5 * time.Minute

func Schedule(ctx context.Context) {
	ctx, span := telemetry.Start(ctx, "batch.Schedule")
	defer span.End()

	ctx = context.WithValue(ctx, types.ContextSecurityKey, types.ContextSecurity{
		Owner: "BatchOperations",
	})

	ticker := time.NewTicker(config.Get().GetBatchShedulePeriod())

	for ctx.Err() == nil {
		go func() {
			ctx, cancel := context.WithTimeout(ctx, config.Get().GetBatchShedulePeriod())
			defer cancel()

			ctx, span := telemetry.Start(ctx, "batch.scheduleBatch", trace.WithNewRoot())
			defer span.End()

			if err := Execute(ctx); err != nil {
				log.WithError(err).Error()
			}
		}()

		select {
		case <-ticker.C:
		case <-ctx.Done():
		}
	}
}

func scaleDownALL(ctx context.Context) error {
	ctx, span := telemetry.Start(ctx, "api.scaleDownALL")
	defer span.End()

	environments, err := api.GetEnvironments(ctx, "")
	if err != nil {
		return errors.Wrap(err, "error listing environments")
	}

	for _, environment := range environments {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := func(environment *api.Environment) error {
			// iteration must have own context
			ctx, cancel := context.WithTimeout(ctx, maxScaleDownDuration)
			defer cancel()

			log := log.WithField("namespace", environment.Namespace)

			// get latest annotations and labels
			err := environment.ReloadFromNamespace(ctx)
			if err != nil {
				return errors.Wrap(err, "error reload environment")
			}

			if environment.NeedToScaleDown(time.Now(), 1) {
				eventMessage := environment.NewWebhookMessage(types.EventPrestop)
				eventMessage.Reason = "Will be scaled down soon..."
				eventMessage.Properties["slackEmoji"] = ":warning:"

				if err := webhook.NewEvent(ctx, eventMessage); err != nil {
					log.WithError(err).Error()
				}
			}

			if !environment.NeedToScaleDown(time.Now(), 0) {
				return nil
			}

			log.Info("scaledown")

			err = environment.ScaleALL(ctx, 0)
			if err != nil {
				return errors.Wrap(err, "error scale down")
			}

			return nil
		}(environment)
		if err != nil {
			log.WithError(err).Error()
		}
	}

	// scaledown servers
	servers, err := api.GetRemoteServers(ctx)
	if err != nil {
		return errors.Wrap(err, "error listing servers")
	}

	for _, server := range servers {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := func(server *api.GetRemoteServerItem) error {
			// iteration must have own context
			ctx, cancel := context.WithTimeout(ctx, maxScaleDownDuration)
			defer cancel()

			log := log.WithField("server", server.Name)
			// calculate is delay is active
			if delay, ok := server.Labels[config.LabelScaleDownDelayShort]; ok {
				scaleDelayTime, err := utils.UnixToTime(delay)
				if err != nil {
					return errors.Wrap(err, "error parse scale delay time")
				} else if time.Now().Before(scaleDelayTime) {
					log.Info("scale down delay is active")

					return nil
				}
			}

			log.Info("scaledown server")

			err := api.SetRemoteServerAction(ctx, api.SetRemoteServerActionInput{
				Cloud:  server.Cloud,
				ID:     server.ID,
				Action: api.SetRemoteServerStatusPowerOff,
			})
			if err != nil {
				return errors.Wrapf(err, "error power off server %s", server.ID)
			}

			return nil
		}(server)
		if err != nil {
			log.WithError(err).Error()
		}
	}

	return nil
}

func Execute(ctx context.Context) error {
	ctx, span := telemetry.Start(ctx, "api.batch")
	defer span.End()

	if err := scaleDownALL(ctx); err != nil {
		log.WithError(err).Error()
	}

	environments, err := api.GetEnvironments(ctx, "")
	if err != nil {
		return errors.Wrap(err, "error list ingress")
	}

	for _, environment := range environments {
		if ctx.Err() != nil {
			return ctx.Err()
		}

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

		reason, description := environment.IsStaled(0)

		log.WithField("reason", reason).Debug(description)

		if reason != api.StaledReasonNone {
			deleteALLResult := environment.DeleteALL(ctx)

			log.Info(deleteALLResult.JSON())
		}
	}

	return nil
}
