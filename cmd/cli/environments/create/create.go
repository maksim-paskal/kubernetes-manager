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
	"strings"

	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/global"
	apiclient "github.com/maksim-paskal/kubernetes-manager/internal/client"
	"github.com/spf13/cobra"
)

// services collects the repeated --service flag, it is package level because the command is.
var services serviceRefs

// serviceRefs is the value of the --service flag, each use adds one <projectID>:<ref>.
type serviceRefs []apiclient.ServiceRef

func (s *serviceRefs) String() string {
	items := make([]string, len(*s))

	for i, service := range *s {
		items[i] = service.String()
	}

	return strings.Join(items, ",")
}

func (s *serviceRefs) Set(value string) error {
	service, err := apiclient.ParseServiceRef(value)
	if err != nil {
		return err
	}

	*s = append(*s, service)

	return nil
}

func (s *serviceRefs) Type() string {
	return "projectID:ref"
}

func Command() *cobra.Command {
	command.Flags().Var(&services, "service", "Service to include in the environment, e.g. 12:main, can be repeated, see 'services list'")
	command.Flags().String("name", "", "Name of the environment")
	command.Flags().String("cluster", "", "Cluster to use for the environment")

	return command
}

var command = &cobra.Command{
	Use:     "create <profile>",
	Short:   "Create environment",
	Long:    "Create environment from the profile, prints the id of the new environment.",
	Example: "  kubernetes-manager environments create my-profile --name feature-x --service 12:main --service 15:feature-x",
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

		global.Output(cmd, id)

		return nil
	},
}
