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
package start

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
	Use:   "start <id>",
	Short: "Start environment",
	Long:  "Start environments, the server starts the pods in background, use --wait to wait until all pods are ready.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := global.Client(cmd)
		if err != nil {
			return err
		}

		return global.RequestAndWait(cmd, global.DefaultWaitTimeout, args,
			func(id string) error {
				if err := client.StartEnvironment(cmd.Context(), id); err != nil {
					return err
				}

				global.Output(cmd, "Environment "+id+" start requested")

				return nil
			},
			func(ctx context.Context, id string) error {
				if err := client.WaitEnvironmentStarted(ctx, id); err != nil {
					return err
				}

				global.Output(cmd, "Environment "+id+" started")

				return nil
			},
		)
	},
}
