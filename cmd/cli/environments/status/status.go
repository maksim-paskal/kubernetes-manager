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
package status

import (
	"strconv"
	"strings"

	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/global"
	apiclient "github.com/maksim-paskal/kubernetes-manager/internal/client"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return command
}

var command = &cobra.Command{
	Use:   "status <id>",
	Short: "Show state of environment pods",
	Long: "Show the state of the environments: paused (no pods), starting, running (all pods are ready), " +
		"failed or stopping (pods are terminating). The pod state can be about 10 seconds old, the server caches it. " +
		"It does not tell whether the services are deployed, see environments services.",
	Example: "  kubernetes-manager environments status cluster:namespace",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := global.Client(cmd)
		if err != nil {
			return err
		}

		statuses := make([]*apiclient.EnvironmentStatus, 0, len(args))

		// the failed ids are printed by ForEachID, the table still shows the rest
		forEachErr := global.ForEachID(cmd, args, apiclient.ValidateEnvironmentID, func(id string) error {
			status, err := client.GetEnvironmentStatus(cmd.Context(), id)
			if err != nil {
				return err
			}

			statuses = append(statuses, status)

			return nil
		})

		if forEachErr != nil && len(statuses) == 0 {
			return forEachErr
		}

		rows := make([][]string, len(statuses))
		for i, status := range statuses {
			rows[i] = []string{
				status.ID,
				status.State,
				strconv.FormatInt(status.Pods.PodsReady, 10) + "/" + strconv.FormatInt(status.Pods.PodsTotal, 10),
				strings.Join(status.Pods.PodsFailedName, ", "),
			}
		}

		global.OutputTable(cmd, statuses, []string{"ID", "STATE", "READY", "FAILED PODS"}, rows)

		return forEachErr
	},
}
