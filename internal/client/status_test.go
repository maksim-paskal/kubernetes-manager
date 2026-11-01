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
	"testing"

	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
)

func TestPodsState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		pods types.PodsInfo
		want string
	}{
		{name: "no pods", pods: types.PodsInfo{}, want: "paused"},
		{name: "terminating", pods: types.PodsInfo{PodsTotal: 2, PodsReady: 1, PodsTerminating: 1}, want: "stopping"},
		{name: "failed", pods: types.PodsInfo{PodsTotal: 2, PodsReady: 1, PodsFailed: 1}, want: "failed"},
		{name: "ready", pods: types.PodsInfo{PodsTotal: 2, PodsReady: 2}, want: "running"},
		{name: "starting", pods: types.PodsInfo{PodsTotal: 2, PodsReady: 1, PodsStarting: 1}, want: "starting"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := podsState(&tt.pods); got != tt.want {
				t.Fatalf("state = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetEnvironmentStatus(t *testing.T) {
	t.Parallel()

	cli, _ := podsServer(t, []*types.PodsInfo{{PodsTotal: 2, PodsReady: 2}})

	status, err := cli.GetEnvironmentStatus(context.Background(), "cluster:namespace")
	if err != nil {
		t.Fatal(err)
	}

	if status.ID != "cluster:namespace" || status.State != "running" || status.Pods.PodsReady != 2 {
		t.Fatalf("status = %+v", status)
	}
}
