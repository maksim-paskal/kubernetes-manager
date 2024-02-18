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
	"fmt"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
)

type EnvironmentBadge struct {
	Key         string
	Value       string
	Description string
}

const (
	stalledDaysToSimmulate        = 3
	needToScaleDownHoursSimmulate = 2
	badgeLastStarted              = "Last started"
)

func (e *Environment) getBadges() []*EnvironmentBadge {
	result := make([]*EnvironmentBadge, 0)

	if len(e.NamespaceCreatedBy) > 0 {
		result = append(result, &EnvironmentBadge{
			Key:   "Created by",
			Value: e.NamespaceCreatedBy,
		})
	}

	if len(e.NamespaceLastScaled) > 0 {
		namespaceLastScaled, _ := utils.StringToTime(e.NamespaceLastScaled)

		text := utils.HumanizeDuration(utils.HumanizeDurationShort, time.Since(namespaceLastScaled))

		result = append(result, &EnvironmentBadge{
			Key:         badgeLastStarted,
			Value:       text + " ago",
			Description: e.NamespaceLastScaled,
		})
	}

	if reason, _ := e.IsStaled(stalledDaysToSimmulate); reason != StaledReasonNone {
		result = append(result, &EnvironmentBadge{
			Key:         "staled",
			Value:       "true",
			Description: "environment will be removed",
		})
	}

	if e.NeedToScaleDown(time.Now(), needToScaleDownHoursSimmulate) {
		result = append(result, &EnvironmentBadge{
			Key:         "scaledown",
			Value:       "true",
			Description: fmt.Sprintf("environment will be scaled down in %d hours", needToScaleDownHoursSimmulate),
		})
	}

	return result
}
