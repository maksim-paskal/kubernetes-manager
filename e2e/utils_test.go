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
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/api"
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/maksim-paskal/kubernetes-manager/pkg/utils"
)

const (
	ID           = "default:test-kubernetes-manager"
	PODLABELS    = "app=envoy-control-plane"
	PODCONTAINER = "envoy-control-plane"
)

var errWaitTimeExpired = errors.New("time expired")

//nolint: goerr113
func checkLastScaleDate(environment *api.Environment) error {
	lastScaleDate := environment.NamespaceAnotations[config.LabelLastScaleDate]
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
func checkEnvironment(environment *api.Environment) error {
	if want := "test-kubernetes-manager"; environment.Namespace != want {
		return fmt.Errorf("want=%s;got=%s", want, environment.Namespace)
	}

	if want := ID; environment.ID != want {
		return fmt.Errorf("want=%s;got=%s", want, environment.ID)
	}

	return nil
}

//nolint: goerr113
func checkHosts(environment *api.Environment) error {
	if len(environment.Hosts) != 1 {
		return errors.New("hosts not found")
	}

	if want := "https://backend-some-feature-branch.yourdomain.com"; environment.Hosts[0] != want {
		return fmt.Errorf("want=%s;got=%s", want, environment.Hosts[0])
	}

	if len(environment.HostsInternal) != 1 {
		return errors.New("internal hosts not found")
	}

	if want := "https://backend-some-feature-branch-internal.yourdomain.com"; environment.HostsInternal[0] != want {
		return fmt.Errorf("want=%s;got=%s", want, environment.HostsInternal[0])
	}

	return nil
}

func waitForPodCount(environment *api.Environment, count int64) (*api.GetContainersItem, error) {
	total := 0

	for {
		total++

		podInfo, err := environment.GetPodsInfo()
		if err != nil {
			return nil, err
		}

		if podInfo.PodsTotal == count {
			pods, err := environment.GetContainers("", "")
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
