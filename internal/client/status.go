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
package client

import (
	"context"

	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
)

// EnvironmentStatus is the state of the environment, it is derived from the state of its pods.
type EnvironmentStatus struct {
	ID    string
	State string
	Pods  *types.PodsInfo
}

// GetEnvironmentStatus returns the state of the environment.
func (c *Client) GetEnvironmentStatus(ctx context.Context, id string) (*EnvironmentStatus, error) {
	pods, err := c.GetPodsInfo(ctx, id)
	if err != nil {
		return nil, err
	}

	return &EnvironmentStatus{ID: id, State: podsState(pods), Pods: pods}, nil
}

// podsState names the state of the environment from its pods, the checks go from the most to the least specific.
// An environment without pods is paused, a freshly created one stays paused until its services are deployed.
func podsState(pods *types.PodsInfo) string {
	switch {
	case pods.PodsTotal == 0:
		return "paused"
	case pods.PodsTerminating > 0:
		return "stopping"
	case pods.PodsFailed > 0:
		return "failed"
	case pods.PodsReady == pods.PodsTotal:
		return "running"
	default:
		return "starting"
	}
}
