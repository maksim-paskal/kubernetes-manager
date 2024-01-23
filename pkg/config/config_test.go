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
	"context"
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

	ctx := context.TODO()

	test1 := config.GetNamespaceMeta(ctx, "").Labels

	test1["test1"] = "test1"
	if test1["environment"] != "dev" {
		t.Fatal("test1 must have environment variable")
	}

	test2 := config.GetNamespaceMeta(ctx, "").Labels

	test2["test2"] = "test2"
	if test2["test1"] == "test1" {
		t.Fatal("test2 should not have values of test1")
	}

	meta := config.NamespaceMeta{
		Labels: map[string]string{
			"test1": "{{ .WebListen }}",
			"aaa":   "bbb",
		},
		Annotations: map[string]string{
			"test2":      "{{ .WebListen }}",
			"vvv":        "aaaa",
			"someRandom": `{{ RandomSliceElement (list "first" "second" "third" ) }}`,
		},
	}

	metaFormated := meta.GetTemplatedValue(context.TODO())

	if metaFormated.Labels["test1"] != *config.Get().WebListen {
		t.Fatalf("annotation has wrong value %s", metaFormated.Labels["test1"])
	}

	if metaFormated.Annotations["test2"] != *config.Get().WebListen {
		t.Fatalf("annotation has wrong value %s", metaFormated.Annotations["test2"])
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
