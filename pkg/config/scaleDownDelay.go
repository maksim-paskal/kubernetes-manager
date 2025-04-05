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
package config

import (
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
	"github.com/pkg/errors"
)

type ScaleDownDelay struct {
	targetScaleDownTime string
	timezone            string

	minDuration    time.Duration
	maxDuration    time.Duration
	scaleDownDelay time.Time
}

type ScaleDownDelayOpts struct {
	TargetScaleDownTime string
	Timezone            string
}

func NewScaleDownDelay(now time.Time, opts *ScaleDownDelayOpts) (*ScaleDownDelay, error) {
	s := ScaleDownDelay{
		targetScaleDownTime: opts.TargetScaleDownTime,
		timezone:            opts.Timezone,

		minDuration: 3 * time.Hour,  //nolint:mnd
		maxDuration: 14 * time.Hour, //nolint:mnd
	}

	if err := s.setTargetScaleDownTime(now); err != nil {
		return nil, errors.Wrap(err, "failed to set target scale down time")
	}

	// If scaleDownDelay is in the past, or less than 3h, set it to minimalDuration
	if now.After(s.scaleDownDelay) || s.scaleDownDelay.Sub(now) < s.minDuration {
		s.scaleDownDelay = now.Add(s.minDuration)
	}

	// If scaleDownDelay is more than 14h, set it to minDuration
	if s.scaleDownDelay.Sub(now) > s.maxDuration {
		s.scaleDownDelay = now.Add(s.minDuration)
	}

	return &s, nil
}

func (s *ScaleDownDelay) setTargetScaleDownTime(now time.Time) error {
	loc, err := time.LoadLocation(s.timezone)
	if err != nil {
		return errors.Wrap(err, "failed to load location")
	}

	targetScaleDown := now.Format("2006-01-02") + "T" + s.targetScaleDownTime

	result, err := time.ParseInLocation("2006-01-02T15:04:05", targetScaleDown, loc)
	if err != nil {
		return errors.Wrap(err, "failed to parse target scale down time")
	}

	s.scaleDownDelay = result

	return nil
}

func (s *ScaleDownDelay) TimeToString() string {
	return utils.TimeToString(s.scaleDownDelay.UTC())
}

func (s *ScaleDownDelay) TimeToUnix() string {
	return utils.TimeToUnix(s.scaleDownDelay.UTC())
}
