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
package autotests

import (
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/environments/autotests/create"
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/environments/autotests/list"
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/environments/autotests/stop"
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/environments/autotests/tests"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	command.AddCommand(list.Command())
	command.AddCommand(tests.Command())
	command.AddCommand(create.Command())
	command.AddCommand(stop.Command())

	return command
}

var command = &cobra.Command{
	Use:   "autotests",
	Short: "Manage autotests in environments",
}
