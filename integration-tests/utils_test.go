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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/api"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
)

const (
	NS           = "default:test-kubernetes-manager"
	PODLABELS    = "app=envoy-control-plane"
	PODCONTAINER = "envoy-control-plane"
)

var errWaitTimeExpired = errors.New("time expired")

//nolint: goerr113
func checkLastScaleDate(ingress *api.GetIngressList) error {
	lastScaleDate := ingress.NamespaceAnotations[config.LabelLastScaleDate]
	if len(lastScaleDate) == 0 {
		return fmt.Errorf("Namespace has no anotations")
	}

	_, err := utils.StringToTime(lastScaleDate)
	if err != nil {
		return fmt.Errorf("LabelLastScaleDate format error")
	}

	return nil
}

//nolint: goerr113
func checkIngress(ingress *api.GetIngressList) error {
	if want := "kubernetes-manager-test"; ingress.IngressName != want {
		return fmt.Errorf("want=%s;got=%s", want, ingress.IngressName)
	}

	if want := NS; ingress.Namespace != want {
		return fmt.Errorf("want=%s;got=%s", want, ingress.Namespace)
	}

	if want := strings.Split(NS, ":")[1]; ingress.NamespaceName != want {
		return fmt.Errorf("want=%s;got=%s", want, ingress.NamespaceName)
	}

	if want := "some-docker-tag"; ingress.IngressAnotations[config.LabelRegistryTag] != want {
		return fmt.Errorf("want=%s;got=%s", want, ingress.IngressAnotations[config.LabelRegistryTag])
	}

	return nil
}

func waitForPodCount(count int) ([]*api.GetPodsItem, error) {
	total := 0

	for {
		total++

		podCount, err := api.GetRunningPodsCount(NS)
		if err != nil {
			return nil, err
		}

		if podCount == count {
			pods, err := api.GetPods(NS)
			if err != nil {
				return nil, err
			}

			return pods, nil
		}

		if total > 100 {
			return nil, errWaitTimeExpired
		}

		time.Sleep(time.Second)
	}
}
