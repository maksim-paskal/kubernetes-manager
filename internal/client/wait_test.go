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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
)

// podsServer answers the pods requests with the steps in order, the last step is repeated.
// A nil step is a failing request.
func podsServer(t *testing.T, steps []*types.PodsInfo) (*Client, *atomic.Int32) {
	t.Helper()

	var requests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/cluster:namespace/pods" {
			http.NotFound(w, r)

			return
		}

		step := min(int(requests.Add(1))-1, len(steps)-1)

		if steps[step] == nil {
			http.Error(w, "boom", http.StatusInternalServerError)

			return
		}

		_ = json.NewEncoder(w).Encode(map[string]any{"Result": steps[step]})
	}))
	t.Cleanup(server.Close)

	cli := NewClient().WithBaseURL(server.URL).WithBasicAuth("login", "password")
	cli.PollInterval = 5 * time.Millisecond

	return cli, &requests
}

func TestWaitEnvironmentStarted(t *testing.T) {
	t.Parallel()

	cli, requests := podsServer(t, []*types.PodsInfo{
		{},
		{PodsTotal: 2, PodsReady: 1, PodsStarting: 1},
		{PodsTotal: 2, PodsReady: 2},
	})

	if err := cli.WaitEnvironmentStarted(context.Background(), "cluster:namespace"); err != nil {
		t.Fatal(err)
	}

	if got := requests.Load(); got != 3 {
		t.Fatalf("requests = %d, want 3", got)
	}
}

func TestWaitEnvironmentPaused(t *testing.T) {
	t.Parallel()

	cli, requests := podsServer(t, []*types.PodsInfo{
		{PodsTotal: 2, PodsReady: 2},
		{PodsTotal: 1, PodsTerminating: 1},
		{},
	})

	if err := cli.WaitEnvironmentPaused(context.Background(), "cluster:namespace"); err != nil {
		t.Fatal(err)
	}

	if got := requests.Load(); got != 3 {
		t.Fatalf("requests = %d, want 3", got)
	}
}

func TestWaitEnvironmentTimeout(t *testing.T) {
	t.Parallel()

	cli, _ := podsServer(t, []*types.PodsInfo{
		{PodsTotal: 2, PodsReady: 1, PodsFailed: 1, PodsFailedName: []string{"api-0 (CrashLoopBackOff) 6m0s"}},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := cli.WaitEnvironmentStarted(ctx, "cluster:namespace")
	if err == nil {
		t.Fatal("want a timeout error")
	}

	for _, want := range []string{"timed out", "2 total", "1 failed", "api-0 (CrashLoopBackOff)"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not contain %q", err, want)
		}
	}
}

func TestWaitEnvironmentCanceled(t *testing.T) {
	t.Parallel()

	cli, _ := podsServer(t, []*types.PodsInfo{{PodsTotal: 1, PodsStarting: 1}})

	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(50*time.Millisecond, cancel)

	err := cli.WaitEnvironmentStarted(ctx, "cluster:namespace")
	if err == nil || strings.Contains(err.Error(), "timed out") {
		t.Fatalf("error = %v, want a cancel error", err)
	}
}

func TestWaitEnvironmentFailures(t *testing.T) {
	t.Parallel()

	t.Run("recovers", func(t *testing.T) {
		t.Parallel()

		cli, _ := podsServer(t, []*types.PodsInfo{nil, nil, {PodsTotal: 1, PodsReady: 1}})

		if err := cli.WaitEnvironmentStarted(context.Background(), "cluster:namespace"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("gives up", func(t *testing.T) {
		t.Parallel()

		cli, requests := podsServer(t, []*types.PodsInfo{nil})

		err := cli.WaitEnvironmentStarted(context.Background(), "cluster:namespace")
		if err == nil || !strings.Contains(err.Error(), "500") {
			t.Fatalf("error = %v, want the server error", err)
		}

		if got := requests.Load(); got != maxPollFailures {
			t.Fatalf("requests = %d, want %d", got, maxPollFailures)
		}
	})
}

func TestWaitEnvironmentInvalidID(t *testing.T) {
	t.Parallel()

	cli, requests := podsServer(t, []*types.PodsInfo{{}})

	if err := cli.WaitEnvironmentPaused(context.Background(), "nope"); err == nil {
		t.Fatal("want an error for an invalid id")
	}

	if got := requests.Load(); got != 0 {
		t.Fatalf("requests = %d, want 0", got)
	}
}
