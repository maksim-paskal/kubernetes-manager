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
	"strconv"

	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/global"
	apiclient "github.com/maksim-paskal/kubernetes-manager/internal/client"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	command.Flags().Bool("required", false, "Filter by required services")

	return command
}

var command = &cobra.Command{
	Use:     "list <profile>",
	Short:   "List services",
	Aliases: []string{"ls"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := global.Client(cmd)
		if err != nil {
			return err
		}

		services, err := client.GetServices(cmd.Context(), args[0])
		if err != nil {
			return err
		}

		required, _ := cmd.Flags().GetBool("required")
		rows := make([][]string, 0)
		filtered := make([]apiclient.Service, 0, len(services))

		for _, service := range services {
			if required && !service.Required {
				continue
			}

			filtered = append(filtered, service)

			rows = append(rows, []string{
				strconv.FormatInt(service.ProjectID, 10),
				strconv.FormatBool(service.Required),
				service.Description,
			})
		}

		global.OutputTable(cmd,
			filtered,
			[]string{"PROJECT ID", "REQUIRED", "DESCRIPTION"},
			rows,
		)

		return nil
	},
}
