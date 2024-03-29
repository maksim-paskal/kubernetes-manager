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
	"fmt"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/telemetry"
	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/maksim-paskal/kubernetes-manager/pkg/webhook"
	log "github.com/sirupsen/logrus"
)

func (e *Environment) ScaleDownDelay(ctx context.Context, durationTime time.Duration) error {
	ctx, span := telemetry.Start(ctx, "api.ScaleDownDelay")
	defer span.End()

	annotation := e.NamespaceAnnotations
	if annotation == nil {
		annotation = make(map[string]string)
	}

	annotation[config.LabelScaleDownDelay] = utils.TimeToString(time.Now().Add(durationTime))

	err := e.SaveNamespaceMeta(ctx, annotation, e.NamespaceLabels)
	if err != nil {
		return err
	}

	eventMessage := e.NewWebhookMessage(types.EventScaledownDelayed)
	eventMessage.Reason = fmt.Sprintf("Scaledown delayed for %s ...", durationTime.String())
	eventMessage.Properties["slackEmoji"] = ":calendar:"

	err = webhook.NewEvent(ctx, eventMessage)
	if err != nil {
		log.WithError(err).Error("error while sending webhook")
	}

	return nil
}
