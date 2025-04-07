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
package config_test

import (
	"testing"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
)

func TestScaleDownDelay(t *testing.T) {
	t.Parallel()

	opts := config.NewScaleDownDelayOpts()

	opts.Timezone = "Europe/Kyiv"

	tests := map[string]string{
		"2025-01-01T06:00:00Z": "2025-01-01T18:00:00Z",
		"2025-01-01T07:00:00Z": "2025-01-01T18:00:00Z",
		"2025-01-01T08:00:00Z": "2025-01-01T18:00:00Z",
		"2025-01-01T09:00:00Z": "2025-01-01T18:00:00Z",
		"2025-01-01T13:00:00Z": "2025-01-01T18:00:00Z",
		"2025-01-01T15:00:00Z": "2025-01-01T18:00:00Z",
		"2025-01-01T18:00:00Z": "2025-01-01T21:00:00Z", // + 3 hours
		"2025-01-01T19:00:00Z": "2025-01-01T22:00:00Z", // + 3 hours
		"2025-01-01T20:00:00Z": "2025-01-01T23:00:00Z", // + 3 hours
		"2025-01-01T21:00:00Z": "2025-01-02T00:00:00Z", // + 3 hours
		"2025-01-01T22:00:00Z": "2025-01-02T01:00:00Z", // + 3 hours
		"2025-01-02T01:00:00Z": "2025-01-02T04:00:00Z", // + 3 hours
		"2025-01-02T03:00:00Z": "2025-01-02T06:00:00Z", // + 3 hours
	}

	for input, expected := range tests {
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			someTime, err := time.Parse(time.RFC3339, input)
			if err != nil {
				t.Fatal(err)
			}

			scaleDownDelay, err := config.NewScaleDownDelay(someTime, opts)
			if err != nil {
				t.Fatal(err)
			}

			if got := scaleDownDelay.TimeToString(); got != expected {
				t.Fatalf("want=%s,got=%s", expected, got)
			}
		})
	}
}
