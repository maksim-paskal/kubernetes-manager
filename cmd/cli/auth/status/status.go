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
	"fmt"

	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/global"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return command
}

var command = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status with the kubernetes manager",
	RunE: func(cmd *cobra.Command, _ []string) error {
		client, err := global.Client(cmd)
		if err != nil {
			return err
		}

		user, err := client.GetUser(cmd.Context())
		if err != nil {
			return err
		}

		if jsonOutput, _ := cmd.Flags().GetBool("json"); jsonOutput {
			global.Output(cmd, user)

			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "User %s is authenticated\n", user.Name)

		return nil
	},
}
