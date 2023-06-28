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
	"encoding/json"
	"strings"
	"testing"

	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	if err := config.Load(); err != nil {
		t.Fatal(err)
	}

	if want := ":3333"; *config.Get().WebListen != want {
		t.Fatalf("webListen != %s", want)
	}

	links := config.Get().KubernetesEndpoints[0].Links

	linksJSON, err := json.Marshal(&links)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(linksJSON), "\"\"") {
		t.Log(config.Get().String())
		t.Fatal("links in kubernetesendpoints should not contain empty string " + string(linksJSON))
	}

	test1 := config.GetNamespaceMeta("").Labels

	test1["test1"] = "test1"
	if test1["environment"] != "dev" {
		t.Fatal("test1 must have environment variable")
	}

	test2 := config.GetNamespaceMeta("").Labels

	test2["test2"] = "test2"
	if test2["test1"] == "test1" {
		t.Fatal("test2 should not have values of test1")
	}
}

func TestFormatedLinks(t *testing.T) {
	t.Parallel()

	testLink := config.Links{
		LogsURL: "[__Namespace__]",
	}

	formatedLinks, err := testLink.FormatedLinks("test")
	if err != nil {
		t.Fatal(err)
	}

	if want := "[test]"; formatedLinks.LogsURL != want {
		t.Fatalf("want=%s,got=%s", want, formatedLinks.LogsURL)
	}
}
