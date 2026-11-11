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
package list

import (
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/global"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	command.Flags().Bool("all", false, "Show all remote servers")

	return command
}

var command = &cobra.Command{
	Use:     "list",
	Short:   "List remote servers",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, _ []string) error {
		client, err := global.Client(cmd)
		if err != nil {
			return err
		}

		all, _ := cmd.Flags().GetBool("all")

		servers, err := client.GetRemoteServers(cmd.Context(), all)
		if err != nil {
			return err
		}

		rows := make([][]string, len(servers))
		for i, server := range servers {
			rows[i] = []string{server.ID, server.Name, string(server.Status)}
		}

		global.OutputTable(cmd,
			servers,
			[]string{"ID", "NAME", "STATUS"},
			rows,
		)

		return nil
	},
}
