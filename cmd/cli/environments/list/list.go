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
	"github.com/maksim-paskal/kubernetes-manager/pkg/config"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	command.Flags().Bool("all", false, "Show all environments")

	return command
}

var command = &cobra.Command{
	Use:     "list",
	Short:   "List environments",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, _ []string) error {
		client, err := global.Client(cmd)
		if err != nil {
			return err
		}

		params := map[string]string{}

		if all, _ := cmd.Flags().GetBool("all"); !all {
			user, err := client.GetUser(cmd.Context())
			if err != nil {
				return err
			}

			params["filter"] = config.LabelNamespaceCreator + "-" + user.Name + "=true"
		}

		environments, err := client.GetEnvironments(cmd.Context(), params)
		if err != nil {
			return err
		}

		rows := make([][]string, len(environments))
		for i, environment := range environments {
			rows[i] = []string{" " + environment.ID + " ", environment.Namespace}
		}

		global.OutputTable(cmd,
			environments,
			[]string{"ID", "NAMESPACE"},
			rows,
		)

		return nil
	},
}
