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
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
)

// check if namespace will be scale down soon.
// if diff > 0 - simulate IsScaleDownSoon.
// return true if need to scale down.
func (e *Environment) NeedToScaleDown(nowDate time.Time, diffHours int) bool {
	now := nowDate

	if diffHours > 0 {
		now = now.Add(time.Duration(diffHours) * time.Hour)
	}

	// if exists annotation with scale down delay
	if scaleDelayText, ok := e.NamespaceAnnotations[config.LabelScaleDownDelay]; ok {
		// date time in correct format
		scaleDelayTime, err := utils.StringToTime(scaleDelayText)
		if err != nil {
			return false
		}

		// for simulation if scaleDelayTime some date in past - remove false
		if diffHours > 0 && time.Now().After(scaleDelayTime) {
			return false
		}

		// if now > scaleDelayTime
		if now.After(scaleDelayTime) {
			return true
		}
	}

	return false
}
