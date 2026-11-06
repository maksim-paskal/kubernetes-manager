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
package deploy

import (
	"context"
	"strconv"

	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/global"
	apiclient "github.com/maksim-paskal/kubernetes-manager/internal/client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// services collects the repeated --service flag, it is package level because the command is.
var services global.ServiceRefs

func Command() *cobra.Command {
	command.Flags().Var(&services, "service", "Service to deploy as projectID:ref, e.g. 12:main, can be repeated, project ids are in 'services list' and 'environments services'")
	command.Flags().Bool("no-build", false, "deploy without the build stage")

	_ = command.MarkFlagRequired("service")

	return command
}

var command = &cobra.Command{
	Use:   "deploy <id>",
	Short: "Deploy services of environment from a branch",
	Long: "Deploy the services of the environment from the refs, the same as the external services tab of the web interface: " +
		"the server sets the ref of every service in the environment and starts a GitLab pipeline for it. " +
		"Use it after a push, the environment does not update itself and a retry of an old pipeline deploys its old commit. " +
		"The pipeline builds the images and then deploys them, --no-build deploys without the build stage. " +
		"The command returns when the pipelines are created, see 'environments services' to know when the services are deployed or failed.",
	Example: "  kubernetes-manager environments deploy cluster:namespace --service 12:fix-login --service 15:main",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := apiclient.ValidateEnvironmentID(args[0]); err != nil {
			return err
		}

		client, err := global.Client(cmd)
		if err != nil {
			return err
		}

		noBuild, _ := cmd.Flags().GetBool("no-build")

		if err := client.DeployServices(cmd.Context(), args[0], services, !noBuild); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return errors.Wrap(err, "the pipelines may be already created, check 'environments services "+args[0]+"' before retrying")
			}

			return err
		}

		global.Output(cmd, "Deploy of "+strconv.Itoa(len(services))+" services requested in "+args[0])

		return nil
	},
}
