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
package global_test

import (
	"testing"

	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/global"
)

func TestServiceRefsSet(t *testing.T) {
	t.Parallel()

	var services global.ServiceRefs

	if err := services.Set("12:main"); err != nil {
		t.Fatal(err)
	}

	if err := services.Set("15:dev"); err != nil {
		t.Fatal(err)
	}

	if err := services.Set("12:other"); err == nil {
		t.Error("the same project twice must fail")
	}

	if err := services.Set("0:main"); err == nil {
		t.Error("a zero project id must fail")
	}

	if got, want := services.String(), "12:main,15:dev"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}

	if len(services) != 2 {
		t.Errorf("len = %d, want 2", len(services))
	}
}
