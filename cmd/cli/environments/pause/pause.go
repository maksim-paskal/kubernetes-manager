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
package pause

import (
	"context"

	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/global"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	global.AddWaitFlags(command, global.DefaultWaitTimeout)

	return command
}

var command = &cobra.Command{
	Use:   "pause <id>",
	Short: "Pause environment",
	Long:  "Pause environments, the server stops the pods in background, use --wait to wait until all pods are gone.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := global.Client(cmd)
		if err != nil {
			return err
		}

		return global.RequestAndWait(cmd, global.DefaultWaitTimeout, args,
			func(id string) error {
				if err := client.PauseEnvironment(cmd.Context(), id); err != nil {
					return err
				}

				global.Output(cmd, "Environment "+id+" pause requested")

				return nil
			},
			func(ctx context.Context, id string) error {
				if err := client.WaitEnvironmentPaused(ctx, id); err != nil {
					return err
				}

				global.Output(cmd, "Environment "+id+" paused")

				return nil
			},
		)
	},
}
