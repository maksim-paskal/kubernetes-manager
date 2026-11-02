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
	"context"
	"time"

	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/global"
	apiclient "github.com/maksim-paskal/kubernetes-manager/internal/client"
	"github.com/spf13/cobra"
)

// services collects the repeated --service flag, it is package level because the command is.
var services global.ServiceRefs

// defaultCreateWaitTimeout is longer than the default of start and pause, the GitLab builds run before the pods start.
const defaultCreateWaitTimeout = 30 * time.Minute

func Command() *cobra.Command {
	command.Flags().Var(&services, "service", "Service to include in the environment, e.g. 12:main, can be repeated, see 'services list'; the profile may fill in the services it requires")
	command.Flags().String("name", "", "Name of the environment, optional, shown as an annotation, does not change the namespace name")
	command.Flags().String("cluster", "", "Cluster to use for the environment, optional, defaults to the first cluster configured on the server, see 'clusters'")
	global.AddWaitFlags(command, defaultCreateWaitTimeout)

	return command
}

var command = &cobra.Command{
	Use:   "create <profile>",
	Short: "Create environment",
	Long: "Create environment from the profile, prints the id of the new environment. " +
		"The pods start after the GitLab pipelines deploy the services, use --wait to wait until they are ready.",
	Example: "  kubernetes-manager environments create my-profile --name feature-x --service 12:main --service 15:feature-x --wait",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := global.Client(cmd)
		if err != nil {
			return err
		}

		id, err := client.CreateEnvironment(cmd.Context(), apiclient.CreateEnvironmentInput{
			Profile:  args[0],
			Services: services,
			Cluster:  cmd.Flag("cluster").Value.String(),
			Name:     cmd.Flag("name").Value.String(),
		})
		if err != nil {
			return err
		}

		// the id is printed before the wait, so it is still known when the wait times out
		global.Output(cmd, id)

		return global.WaitIDs(cmd, []string{id}, func(ctx context.Context, id string) error {
			if err := client.WaitEnvironmentStarted(ctx, id); err != nil {
				return err
			}

			cmd.PrintErrln("Environment " + id + " is ready")

			return nil
		})
	},
}
