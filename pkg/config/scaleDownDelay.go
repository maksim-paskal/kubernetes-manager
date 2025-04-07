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

type ScaleDownDelayStrategy string

const ScaleDownDelayTillTime ScaleDownDelayStrategy = "TillTime"

type ScaleDownDelayOpts struct {
	Strategy            ScaleDownDelayStrategy
	TargetScaleDownTime string
	Timezone            string
	MinDuration         time.Duration
	MaxDuration         time.Duration
}

func NewScaleDownDelayOpts() *ScaleDownDelayOpts {
	return &ScaleDownDelayOpts{
		Strategy:            ScaleDownDelayTillTime,
		TargetScaleDownTime: "20:00:00",
		Timezone:            "UTC",
		MinDuration:         3 * time.Hour,  //nolint:mnd
		MaxDuration:         14 * time.Hour, //nolint:mnd
	}
}

type ScaleDownDelay struct {
	options        *ScaleDownDelayOpts
	scaleDownDelay time.Time
}

func NewScaleDownDelay(now time.Time, opts *ScaleDownDelayOpts) (*ScaleDownDelay, error) {
	s := ScaleDownDelay{
		options: opts,
	}

	if err := s.setTargetScaleDownTime(now); err != nil {
		return nil, errors.Wrap(err, "failed to set target scale down time")
	}

	// If scaleDownDelay is in the past, or less than 3h, set it to minimalDuration
	if now.After(s.scaleDownDelay) || s.scaleDownDelay.Sub(now) < s.options.MinDuration {
		s.scaleDownDelay = now.Add(s.options.MinDuration)
	}

	// If scaleDownDelay is more than 14h, set it to minDuration
	if s.scaleDownDelay.Sub(now) > s.options.MaxDuration {
		s.scaleDownDelay = now.Add(s.options.MinDuration)
	}

	return &s, nil
}

func (s *ScaleDownDelay) setTargetScaleDownTime(now time.Time) error {
	loc, err := time.LoadLocation(s.options.Timezone)
	if err != nil {
		return errors.Wrap(err, "failed to load location")
	}

	targetScaleDown := now.Format("2006-01-02") + "T" + s.options.TargetScaleDownTime

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
