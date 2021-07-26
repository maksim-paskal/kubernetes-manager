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

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
)

func TestIsSystemBranch(t *testing.T) {
	t.Parallel()

	got := utils.IsSystemBranch("master")
	want := true

	if got != want {
		t.Errorf("TestIsSystemBranch, got=%t want=%t", got, want)
	}

	got = utils.IsSystemBranch("test")
	want = false

	if got != want {
		t.Errorf("TestIsSystemBranch, got=%t want=%t", got, want)
	}

	got = utils.IsSystemBranch("test-master1")
	want = false

	if got != want {
		t.Errorf("TestIsSystemBranch, got=%t want=%t", got, want)
	}

	got = utils.IsSystemBranch("release-123456")
	want = true

	if got != want {
		t.Errorf("TestIsSystemBranch, got=%t want=%t", got, want)
	}
}

func TestIsSystemNamespace(t *testing.T) {
	t.Parallel()

	if err := config.Load(); err != nil {
		t.Fatal(err)
	}

	// namespace, isSystemNamespace
	testCases := make(map[string]bool)

	testCases["master"] = false
	testCases["app"] = true
	testCases["release-123456"] = false
	testCases["kube-system-test"] = false

	for namespace, want := range testCases {
		if got := utils.IsSystemNamespace(namespace); got != want {
			t.Errorf("TestIsSystemNamespace, systemNamespaces=%s, namespace=%s, got=%t want=%t",
				*config.Get().SystemNamespaces,
				namespace,
				got,
				want,
			)
		}
	}
}
