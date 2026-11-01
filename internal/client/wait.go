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
	"fmt"
	"strings"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
	"github.com/pkg/errors"
)

// defaultPollInterval is how often the pods are checked, the server caches them for about 10 seconds anyway.
const defaultPollInterval = 5 * time.Second

// maxPollFailures is how many requests in a row can fail while waiting, a failing server must not hang the command.
const maxPollFailures = 3

// GetPodsInfo returns the state of the pods of the environment.
func (c *Client) GetPodsInfo(ctx context.Context, id string) (*types.PodsInfo, error) {
	path, err := envPath(id, "pods")
	if err != nil {
		return nil, err
	}

	result, err := c.get[types.PodsInfo](ctx, path, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pods")
	}

	return &result, nil
}

// WaitEnvironmentStarted waits until all pods of the environment are ready.
// An environment without pods is not started yet, the server creates the pods after the start request.
func (c *Client) WaitEnvironmentStarted(ctx context.Context, id string) error {
	return c.waitPods(ctx, id, func(pods *types.PodsInfo) bool {
		return pods.PodsTotal > 0 && pods.PodsReady == pods.PodsTotal
	})
}

// WaitEnvironmentPaused waits until the environment has no pods, terminating pods are still counted.
func (c *Client) WaitEnvironmentPaused(ctx context.Context, id string) error {
	return c.waitPods(ctx, id, func(pods *types.PodsInfo) bool {
		return pods.PodsTotal == 0
	})
}

// waitPods polls the pods of the environment until done returns true, the ctx deadline is the timeout.
func (c *Client) waitPods(ctx context.Context, id string, done func(*types.PodsInfo) bool) error {
	interval := c.PollInterval
	if interval <= 0 {
		interval = defaultPollInterval
	}

	var last *types.PodsInfo

	failures := 0

	for {
		pods, err := c.GetPodsInfo(ctx, id)

		switch {
		case err == nil:
			failures = 0
			last = pods

			if done(pods) {
				return nil
			}
		case ctx.Err() == nil:
			failures++

			if failures >= maxPollFailures {
				return err
			}
		}

		select {
		case <-ctx.Done():
			return waitError(ctx, last)
		case <-time.After(interval):
		}
	}
}

// waitError explains why the wait stopped, with the last known state of the pods.
func waitError(ctx context.Context, last *types.PodsInfo) error {
	if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return ctx.Err()
	}

	if last == nil {
		return errors.New("timed out waiting for the environment")
	}

	message := fmt.Sprintf("timed out waiting for the environment, pods: %d total, %d ready, %d starting, %d failed, %d terminating",
		last.PodsTotal, last.PodsReady, last.PodsStarting, last.PodsFailed, last.PodsTerminating)

	if len(last.PodsFailedName) > 0 {
		message += ": " + strings.Join(last.PodsFailedName, ", ")
	}

	return errors.New(message)
}
