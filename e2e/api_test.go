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
	"os"
	"sync"
	"testing"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/api"
	"github.com/maksim-paskal/kubernetes-manager/pkg/cache"
	"github.com/maksim-paskal/kubernetes-manager/pkg/client"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
)

const (
	TotalTestCount      = 5
	TestNamespaceFilter = "test-kubernetes-manager=true"
)

var ctx = context.Background()

var counters sync.Map

func init() { //nolint:gochecknoinits
	err := config.Load()
	if err != nil {
		panic(err)
	}

	err = cache.Init(ctx, cache.NoopProvider, nil)
	if err != nil {
		panic(err)
	}

	err = client.Init()
	if err != nil {
		panic(err)
	}
}

func TestEnvironment(t *testing.T) {
	defer counters.Store("TestEnvironment", "Done")

	t.Parallel()

	envieronments, err := api.GetEnvironments(ctx, TestNamespaceFilter)
	if err != nil {
		t.Fatal(err)
	}

	if len(envieronments) != 1 {
		t.Fatal("envieronment not found")
	}

	err = checkEnvironment(envieronments[0])
	if err != nil {
		t.Fatal(err)
	}

	err = checkHosts(envieronments[0])
	if err != nil {
		t.Fatal(err)
	}
}

func TestPods(t *testing.T) {
	defer counters.Store("TestPods", "Done")

	t.Parallel()

	environment, err := api.GetEnvironmentByID(ctx, ID)
	if err != nil {
		t.Fatal(err)
	}

	err = environment.ScaleNamespace(ctx, 2)
	if err != nil {
		t.Fatal(err)
	}

	err = environment.ReloadFromNamespace(ctx)
	if err != nil {
		t.Fatal(err)
	}

	annotation := environment.NamespaceAnnotations

	// after ScaleNamespace, annotation should exists
	if _, ok := annotation[config.LabelLastScaleDate]; !ok {
		t.Fatal("annotation not found")
	}

	annotation["test"] = "value"

	err = environment.SaveNamespaceMeta(ctx, annotation, environment.NamespaceLabels)
	if err != nil {
		t.Fatal(err)
	}

	envieronments, err := api.GetEnvironments(ctx, TestNamespaceFilter)
	if err != nil {
		t.Fatal(err)
	}

	if len(envieronments) != 1 {
		t.Fatal("envieronment not found")
	}

	environment = envieronments[0]

	if environment.NamespaceDescription != os.Getenv("POD_NAMESPACE") {
		t.Fatal("description must equals POD_NAMESPACE env")
	}

	err = checkLastScaleDate(environment)
	if err != nil {
		t.Fatal(err)
	}

	_, err = waitForPodCount(environment, 2)
	if err != nil {
		t.Fatal(err)
	}

	err = environment.DisableHPA(ctx)
	if err != nil {
		t.Fatal(err)
	}

	containers, err := waitForPodCount(environment, 1)
	if err != nil {
		t.Fatal(err)
	}

	alpineImage, err := environment.GetPodByImage(ctx, "alpine:latest")
	if err != nil {
		t.Fatal(err)
	}

	if !alpineImage.Found {
		t.Fatal("alpine image must be found")
	}

	fakeImage, err := environment.GetPodByImage(ctx, "somefakeimage")
	if err != nil {
		t.Fatal(err)
	}

	if fakeImage.Found {
		t.Fatal("fake image must not be found")
	}

	containerName := containers.Contaners[0]

	resultsByPodName, err := environment.ExecContainer(ctx, containerName, "ls")
	if err != nil {
		t.Fatal(err)
	}

	if len(resultsByPodName.ExecCode) != 0 {
		t.Fatal("exit code must be empty")
	}

	resultsByPodLabels, err := environment.ExecContainer(ctx, containerName, "ls")
	if err != nil {
		t.Fatal(err)
	}

	if len(resultsByPodLabels.ExecCode) != 0 {
		t.Fatal("exit code must be empty")
	}

	if resultsByPodName.Stdout != resultsByPodLabels.Stdout {
		t.Fatal("results not valid")
	}

	err = environment.DisableMTLS(ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func TestToken(t *testing.T) {
	defer counters.Store("TestToken", "Done")

	t.Parallel()

	environment, err := api.GetEnvironmentByID(ctx, ID)
	if err != nil {
		t.Fatal(err)
	}

	getClusterKubeconfig, err := environment.GetKubeconfig(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if want := "https://some-public-kubernetes-endpoint"; getClusterKubeconfig.Endpoint != want {
		t.Fatal("Endpoint have wrong data")
	}
}

func TestServices(t *testing.T) {
	defer counters.Store("TestServices", "Done")

	t.Parallel()

	environment, err := api.GetEnvironmentByID(ctx, ID)
	if err != nil {
		t.Fatal(err)
	}

	list, err := environment.GetServices(ctx)
	if err != nil {
		t.Fatal(err)
	}

	serviceFound := false

	for _, service := range list {
		if service.ServiceHost == "envoy-control-plane.test-kubernetes-manager.svc.cluster.local" {
			serviceFound = true

			break
		}
	}

	if !serviceFound {
		t.Fatal("service not found")
	}

	if want := "80"; list[0].Ports != want {
		t.Fatal("service has incorrect port")
	}
}

func TestDeleteNamespace(t *testing.T) {
	t.Parallel()

	environment, err := api.GetEnvironmentByID(ctx, ID)
	if err != nil {
		t.Fatal(err)
	}

	for {
		count := 0

		counters.Range(func(_ interface{}, _ interface{}) bool {
			count++

			return true
		})

		// Total compleated test for deleting namespace
		if count >= TotalTestCount {
			break
		}

		time.Sleep(time.Second)
	}

	if err := environment.DeleteNamespace(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteClusterRole(t *testing.T) {
	defer counters.Store("TestDeleteClusterRole", "Done")

	t.Parallel()

	environment, err := api.GetEnvironmentByID(ctx, ID)
	if err != nil {
		t.Fatal(err)
	}

	err = environment.DeleteClusterRolesAndBindings(ctx)
	if err != nil {
		t.Fatal(err)
	}
}
