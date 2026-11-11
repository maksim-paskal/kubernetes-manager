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
package list

import (
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/global"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	command.Flags().Int("size", 10, "Number of runs to show, 0 shows all of them")

	return command
}

var command = &cobra.Command{
	Use:     "list <id>",
	Short:   "List autotest runs in environment",
	Aliases: []string{"ls"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		size, _ := cmd.Flags().GetInt("size")
		if size < 0 {
			return errors.Errorf("size must not be negative, got %d", size)
		}

		client, err := global.Client(cmd)
		if err != nil {
			return err
		}

		pipelines, err := client.GetAutotests(cmd.Context(), args[0], size)
		if err != nil {
			return err
		}

		rows := make([][]string, len(pipelines))
		for i, pipeline := range pipelines {
			rows[i] = []string{
				pipeline.PipelineID,
				pipeline.Test,
				pipeline.Status,
				pipeline.PipelineOwner,
				pipeline.PipelineCreatedHuman,
				pipeline.PipelineDuration,
			}
		}

		global.OutputTable(cmd,
			pipelines,
			[]string{"ID", "TEST", "STATUS", "OWNER", "CREATED", "DURATION"},
			rows,
		)

		return nil
	},
}
