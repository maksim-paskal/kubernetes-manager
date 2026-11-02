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
package services

import (
	"strconv"

	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/global"
	apiclient "github.com/maksim-paskal/kubernetes-manager/internal/client"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return command
}

var command = &cobra.Command{
	Use:     "services <id>",
	Aliases: []string{"srv"},
	Short:   "Show deploy state of environment services",
	Long: "Show the deploy state of the external services of the environment, the same as the external services tab of the web interface. " +
		"Every service has the branch and the state: deployed, deploying (the pipeline is not finished) " +
		"or failed. The state of the environment is failed when any service failed (the services are listed), else deploying when any " +
		"service is deploying, else ready. It is unknown when the server could not check a service and empty without services. " +
		"Only the services of the environment are shown. The state is the data, the command succeeds for a failed environment.",
	Example: "  kubernetes-manager environments services cluster:namespace --json",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := apiclient.ValidateEnvironmentID(args[0]); err != nil {
			return err
		}

		client, err := global.Client(cmd)
		if err != nil {
			return err
		}

		status, err := client.GetServicesStatus(cmd.Context(), args[0])
		if err != nil {
			return err
		}

		rows := make([][]string, len(status.Services))
		for i, service := range status.Services {
			rows[i] = []string{
				strconv.FormatInt(service.ProjectID, 10),
				service.Name,
				service.Branch,
				service.State,
			}
		}

		global.OutputTable(cmd, status, []string{"PROJECT ID", "SERVICE", "BRANCH", "STATE"}, rows)

		return nil
	},
}
