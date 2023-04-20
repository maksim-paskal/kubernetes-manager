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
package utils_test

import (
	"testing"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
)

func TestDiffToNowHours(t *testing.T) {
	t.Parallel()

	threeHoursLater := time.Now().Add(-time.Hour * 3)

	if was := utils.DiffToNowHours(threeHoursLater); was != 3 {
		t.Fatalf("expected 3 hour, got %d", was)
	}
}

func TestDiffToNowDays(t *testing.T) {
	t.Parallel()

	fiveDaysLater := time.Now().Add(-time.Hour * 5 * 24)

	if was := utils.DiffToNowDays(fiveDaysLater); was != 5 {
		t.Fatalf("expected 5 days, got %d", was)
	}
}

const (
	testTime = "2018-01-01T01:12:34Z"
)

func TestTimeToString(t *testing.T) {
	t.Parallel()

	test, err := time.Parse(time.RFC3339, testTime)
	if err != nil {
		t.Fatal(err)
	}

	if was := utils.TimeToString(test); was != testTime {
		t.Fatalf("expected %s, got %s", testTime, was)
	}
}

func TestStringToTime(t *testing.T) {
	t.Parallel()

	test, err := time.Parse(time.RFC3339, testTime)
	if err != nil {
		t.Fatal(err)
	}

	testResult, err := utils.StringToTime(testTime)
	if err != nil {
		t.Fatal(err)
	}

	if testResult.UTC().String() != test.UTC().String() {
		t.Fatalf("expected %s, got %s", test.UTC().String(), testResult.UTC().String())
	}
}

func TestConvertStringToInt64(t *testing.T) {
	t.Parallel()

	test := "123"

	result, err := utils.ConvertStringToInt64(test)
	if err != nil {
		t.Fatal(err)
	}

	if result != 123 {
		t.Fatalf("expected 123, got %d", result)
	}
}

func TestGetTemplatedResult(t *testing.T) {
	t.Parallel()

	test := "my {{ .Value }}"

	result, err := utils.GetTemplatedResult(test, struct{ Value string }{Value: "test"})
	if err != nil {
		t.Fatal(err)
	}

	if string(result) != "my test" {
		t.Fatalf("expected 'my test', got %s", result)
	}
}
