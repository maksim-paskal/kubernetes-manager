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
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/api"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func IsScaleDownActive(now time.Time) bool {
	batchSheduleTimezone, err := time.LoadLocation(*config.Get().BatchSheduleTimezone)
	if err != nil {
		log.WithError(err).Fatal()
	}

	timeMin := time.Date(now.Year(), now.Month(), now.Day(), config.ScaleDownHourMinPeriod, 0, 0, 0, batchSheduleTimezone)
	timeMax := time.Date(now.Year(), now.Month(), now.Day(), config.ScaleDownHourMaxPeriod, 0, 0, 0, batchSheduleTimezone)

	if now.After(timeMin) || now.Equal(timeMin) {
		return true
	}

	if now.Before(timeMax) || now.Equal(timeMax) {
		return true
	}

	return false
}

// check if scale down is active, namespace will be scaled if false, there is no scaledown if
// if namespace created less than 60m
// if last scale date less than 60m
// if user ask to nodelay.
func IsScaledownDelay(nowTime time.Time, environment *api.Environment) (bool, error) {
	log := log.WithField("namespace", environment.Namespace)

	if len(environment.NamespaceCreated) > 0 {
		namespaceCreatedTime, err := utils.StringToTime(environment.NamespaceCreated)
		if err != nil {
			return false, errors.Wrap(err, "can not parse namespace created time")
		}

		if scaledownDelay := namespaceCreatedTime.Add(namespaceCreatedDelay); nowTime.Before(scaledownDelay) {
			log.Infof("namespace is created less than %s ago, skip", namespaceCreatedDelay.String())

			return true, nil
		}
	}

	if len(environment.NamespaceLastScaled) > 0 {
		namespaceLastScaledTime, err := utils.StringToTime(environment.NamespaceLastScaled)
		if err != nil {
			return false, errors.Wrap(err, "can not parse namespace last scaled time")
		}

		if scaledownDelay := namespaceLastScaledTime.Add(namespaceLastScaledDelay); nowTime.Before(scaledownDelay) {
			log.Infof("namespace is scaled less than %s ago, skip", namespaceLastScaledDelay.String())

			return true, nil
		}
	}

	if scaleDelayText, ok := environment.NamespaceAnotations[config.LabelScaleDownDelay]; ok {
		scaleDelayTime, err := utils.StringToTime(scaleDelayText)
		if err != nil {
			return false, errors.Wrap(err, "error parsing scale delay time")
		}

		if nowTime.Before(scaleDelayTime) {
			log.Info("scale down delay is active")

			return true, nil
		}
	}

	return false, nil
}
