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
package batch_test

import (
	"testing"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/batch"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
)

func TestIsScaleDownActive(t *testing.T) {
	t.Parallel()

	if err := config.Load(); err != nil {
		t.Fatal(err)
	}

	batchSheduleTimezone, err := time.LoadLocation(*config.Get().BatchSheduleTimezone)
	if err != nil {
		t.Fatal(err)
	}

	tests := make(map[time.Time]bool)

	now := time.Now()

	tests[time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, batchSheduleTimezone)] = true
	tests[time.Date(now.Year(), now.Month(), now.Day(), 1, 0, 0, 0, batchSheduleTimezone)] = true
	tests[time.Date(now.Year(), now.Month(), now.Day(), 5, 0, 0, 0, batchSheduleTimezone)] = true

	tests[time.Date(now.Year(), now.Month(), now.Day(), 5, 30, 0, 0, batchSheduleTimezone)] = false
	tests[time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, batchSheduleTimezone)] = false
	tests[time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, batchSheduleTimezone)] = false
	tests[time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, batchSheduleTimezone)] = false

	tests[time.Date(now.Year(), now.Month(), now.Day(), 19, 0, 0, 0, batchSheduleTimezone)] = true
	tests[time.Date(now.Year(), now.Month(), now.Day(), 19, 30, 0, 0, batchSheduleTimezone)] = true

	for k, v := range tests {
		if actual := batch.IsScaleDownActive(k); actual != v {
			t.Fatalf("%s must be %t, actual=%t", k, v, actual)
		}
	}
}
