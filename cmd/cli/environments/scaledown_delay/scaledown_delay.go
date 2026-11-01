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
package scaledown_delay

import (
	"time"

	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/global"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return command
}

var command = &cobra.Command{
	Use:   "scaledown-delay <id> <delay>",
	Short: "Delay automatic pause of environment",
	Long:  "Delay automatic pause of environment, delay is a duration like 3h or 90m.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		delay, err := time.ParseDuration(args[1])
		if err != nil {
			return errors.Wrapf(err, "invalid delay %q, expected a duration like 3h or 90m", args[1])
		}

		if delay <= 0 {
			return errors.Errorf("delay must be positive, got %s", args[1])
		}

		client, err := global.Client(cmd)
		if err != nil {
			return err
		}

		if err := client.ScaleDownDelayEnvironment(cmd.Context(), args[0], args[1]); err != nil {
			return err
		}

		global.Output(cmd, "Environment "+args[0]+" automatic pause delayed for "+args[1])

		return nil
	},
}
