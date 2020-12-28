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
package main

import (
	"testing"
)

func TestIsSystemBranch(t *testing.T) {
	t.Parallel()

	systemGitTags := "master,release-.*"
	appConfig.systemGitTags = &systemGitTags

	got := isSystemBranch("master")
	want := true

	if got != want {
		t.Errorf("TestIsSystemBranch, got=%t want=%t", got, want)
	}

	got = isSystemBranch("test")
	want = false

	if got != want {
		t.Errorf("TestIsSystemBranch, got=%t want=%t", got, want)
	}

	got = isSystemBranch("release-123456")
	want = true

	if got != want {
		t.Errorf("TestIsSystemBranch, got=%t want=%t", got, want)
	}
}

func TestIsSystemNamespace(t *testing.T) {
	t.Parallel()

	systemNamespaces := "kube-system,app"
	appConfig.systemNamespaces = &systemNamespaces

	got := isSystemNamespace("master")
	want := false

	if got != want {
		t.Errorf("TestIsSystemNamespace, got=%t want=%t", got, want)
	}

	got = isSystemNamespace("app")
	want = true

	if got != want {
		t.Errorf("TestIsSystemNamespace, got=%t want=%t", got, want)
	}

	got = isSystemNamespace("release-123456")
	want = false

	if got != want {
		t.Errorf("TestIsSystemNamespace, got=%t want=%t", got, want)
	}
}
