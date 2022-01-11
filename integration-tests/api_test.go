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
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/api"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
)

const TotalTestCount = 4

var counters sync.Map

func init() { //nolint:gochecknoinits
	err := config.Load()
	if err != nil {
		panic(err)
	}

	err = api.Init()
	if err != nil {
		panic(err)
	}
}

func TestIngress(t *testing.T) {
	defer counters.Store("TestIngress", "Done")

	t.Parallel()

	ingress, err := api.GetIngress()
	if err != nil {
		t.Fatal(err)
	}

	if len(ingress) != 1 {
		t.Fatal("no ingress found")
	}

	err = checkIngress(ingress[0])
	if err != nil {
		t.Fatal(err)
	}
}

func TestPods(t *testing.T) {
	defer counters.Store("TestPods", "Done")

	t.Parallel()

	err := api.ScaleNamespace(NS, 2)
	if err != nil {
		t.Fatal(err)
	}

	ingress, err := api.GetIngress()
	if err != nil {
		t.Fatal(err)
	}

	if len(ingress) != 1 {
		t.Fatal("ingress not found")
	}

	err = checkLastScaleDate(ingress[0])
	if err != nil {
		t.Fatal(err)
	}

	_, err = waitForPodCount(2)
	if err != nil {
		t.Fatal(err)
	}

	err = api.DisableHPA(NS)
	if err != nil {
		t.Fatal(err)
	}

	pods, err := waitForPodCount(1)
	if err != nil {
		t.Fatal(err)
	}

	podName := pods[0].PodName

	resultsByPodName, err := api.ExecContainer(NS, podName, "", PODCONTAINER, "ls")
	if err != nil {
		t.Fatal(err)
	}

	if len(resultsByPodName.ExecCode) != 0 {
		t.Fatal("exit code must be empty")
	}

	resultsByPodLabels, err := api.ExecContainer(NS, "", PODLABELS, PODCONTAINER, "ls")
	if err != nil {
		t.Fatal(err)
	}

	if len(resultsByPodLabels.ExecCode) != 0 {
		t.Fatal("exit code must be empty")
	}

	if resultsByPodName.Stdout != resultsByPodLabels.Stdout {
		t.Fatal("results not valid")
	}

	deletePodName := fmt.Sprintf("%s:%s", podName, PODCONTAINER)

	err = api.DeletePod(NS, deletePodName, "")
	if err != nil {
		t.Fatal(err)
	}

	_, err = waitForPodCount(1)
	if err != nil {
		t.Fatal(err)
	}

	err = api.DisableMTLS(NS)
	if err != nil {
		t.Fatal(err)
	}

	err = api.DeletePod(NS, "", PODLABELS)
	if err != nil {
		t.Fatal(err)
	}
}

func TestToken(t *testing.T) {
	defer counters.Store("TestToken", "Done")

	t.Parallel()

	getClusterKubeconfig, err := api.GetClusterKubeconfig("default")
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

	list, err := api.GetServices(NS)
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

	for {
		count := 0

		counters.Range(func(key interface{}, value interface{}) bool {
			count++

			return true
		})

		// Total compleated test for deleting namespace
		if count >= TotalTestCount {
			break
		}

		time.Sleep(time.Second)
	}

	if err := api.DeleteNamespace(NS); err != nil {
		t.Fatal(err)
	}
}
