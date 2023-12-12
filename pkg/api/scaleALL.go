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
package api

import (
	"context"
	"encoding/json"

	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
	"github.com/maksim-paskal/kubernetes-manager/pkg/webhook"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ScaleALL scale namespace and process webhooks.
func (e *Environment) ScaleALL(ctx context.Context, replicas int32) error {
	ctx, span := telemetry.Start(ctx, "api.ScaleALL")
	defer span.End()

	// if scaleDown do nothing if environment have not pods
	if replicas == 0 {
		podInfo, err := e.GetPodsInfo(ctx)
		if err != nil {
			return errors.Wrap(err, "error while getting pods info")
		}

		if podInfo.PodsTotal == 0 {
			return nil
		}
	}

	processWebhook := make(chan error)
	processScale := make(chan error)

	go func() {
		eventType := types.EventStart
		eventReason := "Starting ..."
		eventEmoji := ":white_check_mark:"

		if replicas == 0 {
			eventType = types.EventStop
			eventReason = "Stopping ..."
			eventEmoji = ":no_entry:"
		}

		eventMessage := e.NewWebhookMessage(eventType)
		eventMessage.Reason = eventReason
		eventMessage.Properties["slackEmoji"] = eventEmoji

		processWebhook <- webhook.NewEvent(ctx, eventMessage)
	}()

	go func() {
		processScale <- e.ScaleNamespace(ctx, replicas)
	}()

	type Result struct {
		ErrProcessWebhook string
		ErrProcessScale   string
	}

	result := Result{}
	hasError := false

	if err := <-processWebhook; err != nil {
		hasError = true
		result.ErrProcessWebhook = err.Error()
	}

	if err := <-processScale; err != nil {
		hasError = true
		result.ErrProcessScale = err.Error()
	}

	if hasError {
		resultText, err := json.Marshal(result)
		if err != nil {
			log.WithError(err).Error("error while marshaling result")
		}

		return errors.New(string(resultText))
	}

	return nil
}
