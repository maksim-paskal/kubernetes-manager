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
package issues

import (
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/global"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return command
}

var command = &cobra.Command{
	Use:   "issues <id>",
	Short: "List Sentry issues of environment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := global.Client(cmd)
		if err != nil {
			return err
		}

		response, err := client.GetIssues(cmd.Context(), args[0])
		if err != nil {
			return err
		}

		rows := make([][]string, len(response.Items))
		for i, issue := range response.Items {
			rows[i] = []string{issue.ShortID, issue.Project, issue.Level, issue.LastSeenShort, issue.Title}
		}

		global.OutputTable(cmd,
			response.Items,
			[]string{"ID", "PROJECT", "LEVEL", "LAST SEEN", "TITLE"},
			rows,
		)

		return nil
	},
}
