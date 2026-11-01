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
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/global"
	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return command
}

var command = &cobra.Command{
	Use:   "start <id>",
	Short: "Start remote server",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := global.Client(cmd)
		if err != nil {
			return err
		}

		return global.ForEachID(cmd, args, nil, func(id string) error {
			server, err := client.GetRemoteServer(cmd.Context(), id)
			if err != nil {
				return err
			}

			_, err = client.RemoteServerAction(cmd.Context(), types.SetRemoteServerActionInput{
				Action: types.SetRemoteServerStatusPowerOn,
				ID:     server.ID,
				Cloud:  server.Cloud,
			})
			if err != nil {
				return err
			}

			global.Output(cmd, "Remote server "+id+" started successfully")

			return nil
		})
	},
}
