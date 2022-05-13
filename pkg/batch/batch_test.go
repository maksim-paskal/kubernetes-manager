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

	"github.com/maksim-paskal/kubernetes-manager/pkg/api"
	"github.com/maksim-paskal/kubernetes-manager/pkg/batch"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
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

func TestIsScaledownDelay(t *testing.T) {
	t.Parallel()

	nowDate, err := utils.StringToTime("2022-05-13T09:00:00Z")
	if err != nil {
		t.Fatal(err)
	}

	tests := make(map[*api.GetIngressList]bool)

	tests[&api.GetIngressList{
		Namespace:        "test-001",
		NamespaceCreated: "2022-05-13T08:10:00Z",
	}] = true // namespace will not scaledown

	tests[&api.GetIngressList{
		Namespace:        "test-002",
		NamespaceCreated: "2022-05-13T07:10:00Z",
	}] = false // namespace will be scaledown

	tests[&api.GetIngressList{
		Namespace:           "test-003",
		NamespaceCreated:    "2022-05-13T07:10:00Z",
		NamespaceLastScaled: "2022-05-13T08:10:00Z",
	}] = true

	tests[&api.GetIngressList{
		Namespace:           "test-004",
		NamespaceCreated:    "2022-05-13T07:10:00Z",
		NamespaceLastScaled: "2022-05-13T07:10:00Z",
	}] = false

	tests[&api.GetIngressList{
		Namespace:           "test-005",
		NamespaceCreated:    "2022-05-13T07:10:00Z",
		NamespaceLastScaled: "2022-05-13T07:10:00Z",
		NamespaceAnotations: map[string]string{
			config.LabelScaleDownDelay: "2022-05-13T10:00:00Z",
		},
	}] = true

	tests[&api.GetIngressList{
		Namespace:           "test-005",
		NamespaceCreated:    "2022-05-13T07:10:00Z",
		NamespaceLastScaled: "2022-05-13T07:10:00Z",
		NamespaceAnotations: map[string]string{
			config.LabelScaleDownDelay: "2022-05-13T7:00:00Z",
		},
	}] = false

	for k, v := range tests {
		actual, err := batch.IsScaledownDelay(nowDate, k)
		if err != nil {
			t.Fatal(err)
		}

		if actual != v {
			t.Fatalf("%s must be %t, actual=%t", k, v, actual)
		}
	}

	// test with errors

	errorsTests := []*api.GetIngressList{
		{
			Namespace:        "test-001",
			NamespaceCreated: "fake-date",
		},
		{
			Namespace:           "test-002",
			NamespaceLastScaled: "fake-date",
		},
		{
			NamespaceAnotations: map[string]string{
				config.LabelScaleDownDelay: "fake-date",
			},
		},
	}

	for _, v := range errorsTests {
		_, err := batch.IsScaledownDelay(nowDate, v)
		if err == nil {
			t.Fatalf("must be error %s", v.String())
		}
	}
}
