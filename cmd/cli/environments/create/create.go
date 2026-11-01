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
package create

import (
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/global"
	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	command.Flags().String("services", "", "Services to include in the environment")
	command.Flags().String("name", "", "Name of the environment")
	command.Flags().String("cluster", "", "Cluster to use for the environment")

	return command
}

var command = &cobra.Command{
	Use:   "create <profile>",
	Short: "Create environment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := global.Client(cmd)
		if err != nil {
			return err
		}

		id, err := client.CreateEnvironment(cmd.Context(), types.StartNewEnvironmentInput{
			Profile:  args[0],
			Services: cmd.Flag("services").Value.String(),
			Cluster:  cmd.Flag("cluster").Value.String(),
			Name:     cmd.Flag("name").Value.String(),
		})
		if err != nil {
			return err
		}

		global.Output(cmd, id)

		return nil
	},
}
