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
	"github.com/maksim-paskal/kubernetes-manager/pkg/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	command.Flags().String("ref", "", "Ref to run the test on, defaults to the ref configured for the test")
	command.Flags().Bool("force", false, "Start the test even if a run of it is already pending or running")

	return command
}

var command = &cobra.Command{
	Use:   "create <id> <test>",
	Short: "Start autotest in environment",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref, _ := cmd.Flags().GetString("ref")
		force, _ := cmd.Flags().GetBool("force")

		client, err := global.Client(cmd)
		if err != nil {
			return err
		}

		actions, err := client.GetAutotestActions(cmd.Context(), args[0])
		if err != nil {
			return err
		}

		if err := validateTest(actions, args[1]); err != nil {
			return err
		}

		message, err := client.StartAutotest(cmd.Context(), args[0], types.StartAutotestInput{
			Ref:   ref,
			Test:  args[1],
			Force: force,
		})
		if err != nil {
			return err
		}

		global.Output(cmd, message)

		return nil
	},
}

func validateTest(actions []*types.AutotestAction, test string) error {
	available := make([]string, len(actions))

	for i, action := range actions {
		if action.Test == test {
			return nil
		}

		available[i] = action.Test
	}

	if len(available) == 0 {
		return errors.New("no tests are configured for this environment")
	}

	return errors.Errorf("test %q is not configured for this environment, available tests: %s", test, strings.Join(available, ", "))
}
