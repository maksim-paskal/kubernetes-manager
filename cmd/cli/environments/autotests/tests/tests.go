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
package tests

import (
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/global"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return command
}

var command = &cobra.Command{
	Use:   "tests <id>",
	Short: "List tests available in environment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := global.Client(cmd)
		if err != nil {
			return err
		}

		actions, err := client.GetAutotestActions(cmd.Context(), args[0])
		if err != nil {
			return err
		}

		rows := make([][]string, len(actions))
		for i, action := range actions {
			rows[i] = []string{action.Test, action.Name, action.Ref}
		}

		global.OutputTable(cmd,
			actions,
			[]string{"TEST", "NAME", "REF"},
			rows,
		)

		return nil
	},
}
