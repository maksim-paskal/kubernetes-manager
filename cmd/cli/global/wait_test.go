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
package global_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/global"
	"github.com/spf13/cobra"
)

func newWaitCommand(t *testing.T, args ...string) *cobra.Command {
	t.Helper()

	cmd := &cobra.Command{Use: "start", RunE: func(*cobra.Command, []string) error { return nil }}
	global.AddWaitFlags(cmd, global.DefaultWaitTimeout)

	if err := cmd.ParseFlags(args); err != nil {
		t.Fatal(err)
	}

	cmd.SetContext(context.Background())

	return cmd
}

func TestRequestAndWait(t *testing.T) {
	t.Parallel()

	ids := []string{"c:one", "c:two", "c:three"}

	tests := []struct {
		name          string
		args          []string
		failRequest   string
		failWait      string
		wantRequested []string
		wantWaited    []string
		wantErr       bool
	}{
		{name: "without wait", wantRequested: ids},
		{name: "with wait", args: []string{"--wait"}, wantRequested: ids, wantWaited: ids},
		{
			name:          "failed request is not waited",
			args:          []string{"--wait"},
			failRequest:   "c:two",
			wantRequested: []string{"c:one", "c:three"},
			wantWaited:    []string{"c:one", "c:three"},
			wantErr:       true,
		},
		{
			name:          "failed wait does not stop the rest",
			args:          []string{"--wait"},
			failWait:      "c:one",
			wantRequested: ids,
			wantWaited:    ids,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var (
				mu                sync.Mutex
				requested, waited []string
			)

			err := global.RequestAndWait(newWaitCommand(t, tt.args...), global.DefaultWaitTimeout, ids,
				func(id string) error {
					if id == tt.failRequest {
						return errors.New("request failed")
					}

					requested = append(requested, id)

					return nil
				},
				func(_ context.Context, id string) error {
					mu.Lock()

					waited = append(waited, id)
					mu.Unlock()

					if id == tt.failWait {
						return errors.New("wait failed")
					}

					return nil
				},
			)

			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if !slices.Equal(requested, tt.wantRequested) {
				t.Errorf("requested = %v, want %v", requested, tt.wantRequested)
			}

			wantWaited := slices.Sorted(slices.Values(tt.wantWaited))

			slices.Sort(waited)

			if !slices.Equal(waited, wantWaited) {
				t.Errorf("waited = %v, want %v", waited, tt.wantWaited)
			}
		})
	}
}

func TestRequestAndWaitInvalidID(t *testing.T) {
	t.Parallel()

	called := false

	err := global.RequestAndWait(newWaitCommand(t, "--wait"), global.DefaultWaitTimeout, []string{"c:one", "nope"},
		func(string) error {
			called = true

			return nil
		},
		func(context.Context, string) error { return nil },
	)
	if err == nil || called {
		t.Fatalf("error = %v, requested = %v, want an error before any request", err, called)
	}
}

func TestRequestAndWaitSharesTimeout(t *testing.T) {
	t.Parallel()

	var (
		mu        sync.Mutex
		deadlines []time.Time
	)

	err := global.RequestAndWait(newWaitCommand(t, "--wait"), time.Hour, []string{"c:one", "c:two"},
		func(string) error { return nil },
		func(ctx context.Context, _ string) error {
			deadline, ok := ctx.Deadline()
			if !ok {
				t.Error("wait context has no deadline")
			}

			mu.Lock()

			deadlines = append(deadlines, deadline)
			mu.Unlock()

			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(deadlines) != 2 || !deadlines[0].Equal(deadlines[1]) {
		t.Fatalf("deadlines = %v, want one shared deadline", deadlines)
	}
}

func TestWaitIDsWithoutWaitFlag(t *testing.T) {
	t.Parallel()

	called := false

	err := global.WaitIDs(newWaitCommand(t), global.DefaultWaitTimeout, []string{"c:one"}, func(context.Context, string) error {
		called = true

		return nil
	})
	if err != nil || called {
		t.Fatalf("error = %v, waited = %v, want no wait without --wait", err, called)
	}
}

func TestWaitIDsReturnsWaitError(t *testing.T) {
	t.Parallel()

	err := global.WaitIDs(newWaitCommand(t, "--wait"), global.DefaultWaitTimeout, []string{"c:one"}, func(context.Context, string) error {
		return errors.New("stuck")
	})
	if err == nil || !strings.Contains(err.Error(), "1 of 1 failed") {
		t.Fatalf("error = %v, want the failure count", err)
	}
}

func TestWaitIDsDefaultTimeoutWithoutCommandTimeout(t *testing.T) {
	t.Parallel()

	var deadline time.Time

	err := global.WaitIDs(newWaitCommand(t, "--wait"), time.Hour, []string{"c:one"}, func(ctx context.Context, _ string) error {
		deadline, _ = ctx.Deadline()

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if left := time.Until(deadline); left < 59*time.Minute || left > time.Hour {
		t.Fatalf("time left = %v, want the default of about one hour", left)
	}
}

func TestWaitIDsKeepsCommandTimeout(t *testing.T) {
	t.Parallel()

	cmd := newWaitCommand(t, "--wait")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd.SetContext(ctx)

	want, _ := ctx.Deadline()

	var got time.Time

	err := global.WaitIDs(cmd, time.Hour, []string{"c:one"}, func(ctx context.Context, _ string) error {
		got, _ = ctx.Deadline()

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if !got.Equal(want) {
		t.Fatalf("deadline = %v, want the deadline of --timeout %v", got, want)
	}
}
