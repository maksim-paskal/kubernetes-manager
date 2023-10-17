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
package api_test

import (
	"testing"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/api"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
)

func TestNeedToScaleDown(t *testing.T) {
	t.Parallel()

	hour00 := time.Now()
	hour01 := hour00.Add(time.Duration(1) * time.Hour)
	hour02 := hour00.Add(time.Duration(2) * time.Hour)
	hour03 := hour00.Add(time.Duration(3) * time.Hour)

	type testCase struct {
		Now           time.Time
		Delay         time.Time
		Diff          int
		WillScaleDown bool
	}

	testCases := []testCase{
		// for expired delay - need to scale down
		{Now: hour00, Delay: hour00, Diff: 0, WillScaleDown: true},
		{Now: hour01, Delay: hour00, Diff: 0, WillScaleDown: true},
		{Now: hour02, Delay: hour00, Diff: 0, WillScaleDown: true},
		{Now: hour03, Delay: hour00, Diff: 0, WillScaleDown: true},
		// now > delay - do not need to scale down
		{Now: hour00, Delay: hour03, Diff: 0, WillScaleDown: false},
		{Now: hour01, Delay: hour03, Diff: 0, WillScaleDown: false},
		{Now: hour02, Delay: hour03, Diff: 0, WillScaleDown: false},
		{Now: hour03, Delay: hour03, Diff: 0, WillScaleDown: true},
		// add 1 hour to now
		{Now: hour00, Delay: hour00, Diff: 0, WillScaleDown: true},
		{Now: hour00, Delay: hour01, Diff: 0, WillScaleDown: false},
		{Now: hour00, Delay: hour02, Diff: 0, WillScaleDown: false},
		{Now: hour00, Delay: hour03, Diff: 0, WillScaleDown: false},
		// add 1 hour to now
		{Now: hour00, Delay: hour00, Diff: 1, WillScaleDown: false},
		{Now: hour00, Delay: hour01, Diff: 1, WillScaleDown: true},
		{Now: hour00, Delay: hour02, Diff: 1, WillScaleDown: false},
		{Now: hour00, Delay: hour03, Diff: 1, WillScaleDown: false},
		// add 2 hour to now
		{Now: hour00, Delay: hour00, Diff: 2, WillScaleDown: false},
		{Now: hour00, Delay: hour01, Diff: 2, WillScaleDown: true},
		{Now: hour00, Delay: hour02, Diff: 2, WillScaleDown: true},
		{Now: hour00, Delay: hour03, Diff: 2, WillScaleDown: false},
	}

	for i, testCase := range testCases {
		environment := &api.Environment{
			NamespaceAnnotations: map[string]string{
				config.LabelScaleDownDelay: utils.TimeToString(testCase.Delay),
			},
		}

		if got := environment.NeedToScaleDown(testCase.Now, testCase.Diff); got != testCase.WillScaleDown {
			t.Errorf("(case %d) got %v, want %v", i, got, testCase.WillScaleDown)
		}
	}
}
