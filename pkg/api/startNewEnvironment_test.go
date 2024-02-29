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
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/maksim-paskal/kubernetes-manager/pkg/api"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
)

func TestValidation(t *testing.T) {
	t.Parallel()

	if err := config.Load(); err != nil {
		t.Fatal(err)
	}

	valid := make([]api.StartNewEnvironmentInput, 0)

	valid = append(valid, api.StartNewEnvironmentInput{
		Profile:  "test",
		Services: "1:test;2:test2;3:test3",

		Cluster: "test",
	})

	valid = append(valid, api.StartNewEnvironmentInput{
		Profile:  "test",
		Services: "1:test",
		Cluster:  "test2",
	})

	ctx := context.WithValue(context.Background(), types.ContextSecurityKey, types.ContextSecurity{Owner: "test"})

	for _, input := range valid {
		if err := input.Validation(ctx); err != nil {
			t.Fatal(err)
		}

		namespace, err := api.GetNamespaceByServices(input.GetProfile(), input.Services)
		if err != nil {
			t.Fatal(err)
		}

		if need := "test-main-test"; namespace != need {
			t.Fatalf("namespace not correct need=%s;got=%s", need, namespace)
		}
	}
}

func TestParseEnvironmentServices(t *testing.T) {
	t.Parallel()

	valid := make(map[string]int)
	notvalid := make(map[string]int)

	valid["1:test1;2:test2;3:test3"] = 3
	valid["1:test1;2:test2;3:test3;4:test4"] = 4

	notvalid["d"] = 0
	notvalid["d:3"] = 0
	notvalid["1:1,1:2"] = 0

	for services, servicesLen := range valid {
		environmentServices, err := api.ParseEnvironmentServices(services, nil)
		if err != nil {
			t.Fatal(err)
		}

		if environmentServicesLen := len(environmentServices); servicesLen != environmentServicesLen {
			t.Fatal("result length not correct")
		}

		for _, environmentService := range environmentServices {
			if environmentService.Ref != fmt.Sprintf("test%d", environmentService.ProjectID) {
				t.Fatalf("wrong ref %s", environmentService.Ref)
			}
		}
	}

	for services := range notvalid {
		_, err := api.ParseEnvironmentServices(services, nil)
		if err == nil {
			t.Fatal("must return error")
		}
	}
}

func TestParseEnvironmentServicesSort(t *testing.T) {
	t.Parallel()

	type test struct {
		services         string
		order            []int
		sortByProjectIDs []string
	}

	tests := []test{
		{
			services: "1:test;2:test2;3:test3",
			order:    []int{1, 2, 3},
		},
		{
			services: "3:test3;1:test;2:test2",
			order:    []int{3, 1, 2},
		},
		{
			services:         "3:test3;1:test;2:test2",
			order:            []int{2, 1, 3},
			sortByProjectIDs: []string{"2", "1"},
		},
		{
			services:         "1:test3;2:test;3:test2",
			order:            []int{2, 1, 3},
			sortByProjectIDs: []string{"2", "1"},
		},
		{
			services:         "3:test3;2:test;1:test2",
			order:            []int{1, 2, 3},
			sortByProjectIDs: []string{"1", "2"},
		},
	}

	for _, test := range tests {
		environmentServices, err := api.ParseEnvironmentServices(test.services, test.sortByProjectIDs)
		if err != nil {
			t.Fatal(err)
		}

		for i, o := range test.order {
			if environmentServices[i].ProjectID != o {
				t.Fatalf("wrong order for project %d", environmentServices[i].ProjectID)
			}
		}
	}
}

func TestGetNamespaceByServices(t *testing.T) {
	t.Parallel()

	services := "1:test;3:test3;2:test2"
	profile := &config.ProjectProfile{
		NamespacePrefix: "my-test-",
		Required:        "2,3",
	}

	namespace, err := api.GetNamespaceByServices(profile, services)
	if err != nil {
		t.Fatal(err)
	}

	if need := "my-test-test2"; !strings.HasPrefix(namespace, need) {
		t.Fatalf("namespace not correct need=%s;got=%s", need, namespace)
	}
}

func TestGetNamespaceByProfile(t *testing.T) {
	t.Parallel()

	profile := &config.ProjectProfile{
		NamespacePrefix: "my-test-",
	}

	namespace := api.GetNamespaceByProfile(profile)
	if need := "my-test-"; !strings.HasPrefix(namespace, need) {
		t.Fatalf("namespace not correct need=%s;got=%s", need, namespace)
	}

	t.Log(namespace)

	if reqLen := 13; len(namespace) != reqLen {
		t.Fatalf("namespace length not correct need=%d;got=%d", reqLen, len(namespace))
	}
}

func TestGetNamespaceByServicesJIRA(t *testing.T) {
	t.Parallel()

	services := "1:test;3:test3;2:TEST-123"
	profile := &config.ProjectProfile{
		NamespacePrefix: "my-test-",
	}

	namespace, err := api.GetNamespaceByServicesJIRA(profile, services)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(namespace)

	if need := "my-test-TEST-123"; namespace != need {
		t.Fatalf("namespace not correct need=%s;got=%s", need, namespace)
	}
}
