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
package global

import (
	"context"
	"sync"
	"time"

	apiclient "github.com/maksim-paskal/kubernetes-manager/internal/client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// DefaultWaitTimeout is the wait without --timeout of the operations that scale environments.
const DefaultWaitTimeout = 10 * time.Minute

// AddWaitFlags adds --wait to a command that starts an operation the server finishes in background.
// The wait is limited by --timeout, without it by defaultTimeout.
func AddWaitFlags(cmd *cobra.Command, defaultTimeout time.Duration) {
	cmd.Flags().Bool("wait", false, "wait until the operation is finished, without --timeout it gives up after "+defaultTimeout.String())
}

// RequestAndWait sends the request for every environment id. With --wait it then waits for every id that was
// requested, all of them share the same deadline, so the operations run in parallel on the server.
func RequestAndWait(cmd *cobra.Command, defaultTimeout time.Duration, ids []string,
	request func(id string) error, wait func(ctx context.Context, id string) error,
) error {
	requested := make([]string, 0, len(ids))

	requestErr := ForEachID(cmd, ids, apiclient.ValidateEnvironmentID, func(id string) error {
		if err := request(id); err != nil {
			return err
		}

		requested = append(requested, id)

		return nil
	})

	waitErr := WaitIDs(cmd, defaultTimeout, requested, wait)

	if requestErr != nil {
		return requestErr
	}

	return waitErr
}

// WaitIDs waits for every id with --wait, it does nothing without the flag. The ids are waited in parallel,
// they share one deadline, so a slow id must not use the time of the next ones and hide that they are ready.
// The deadline of --timeout is used when it is set, otherwise the wait gets defaultTimeout.
func WaitIDs(cmd *cobra.Command, defaultTimeout time.Duration, ids []string, wait func(ctx context.Context, id string) error) error {
	if shouldWait, _ := cmd.Flags().GetBool("wait"); !shouldWait {
		return nil
	}

	ctx := cmd.Context()

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc

		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}

	errs := make([]error, len(ids))

	var wg sync.WaitGroup

	for i, id := range ids {
		wg.Go(func() {
			errs[i] = wait(ctx, id)
		})
	}

	wg.Wait()

	failed := 0

	for i, err := range errs {
		if err != nil {
			failed++

			cmd.PrintErrln(errors.Wrap(err, ids[i]))
		}
	}

	if failed > 0 {
		return errors.Errorf("%d of %d failed", failed, len(ids))
	}

	return nil
}
