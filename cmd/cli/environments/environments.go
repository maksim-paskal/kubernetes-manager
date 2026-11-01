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
package environments

import (
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/environments/autotests"
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/environments/create"
	envDelete "github.com/maksim-paskal/kubernetes-manager/cmd/cli/environments/delete"
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/environments/get"
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/environments/hosts"
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/environments/issues"
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/environments/kubeconfig"
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/environments/list"
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/environments/mtls"
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/environments/pause"
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/environments/scaledown_delay"
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/environments/start"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	command.AddCommand(list.Command())
	command.AddCommand(get.Command())
	command.AddCommand(create.Command())
	command.AddCommand(envDelete.Command())
	command.AddCommand(start.Command())
	command.AddCommand(pause.Command())
	command.AddCommand(mtls.Command())
	command.AddCommand(kubeconfig.Command())
	command.AddCommand(hosts.Command())
	command.AddCommand(issues.Command())
	command.AddCommand(scaledown_delay.Command())
	command.AddCommand(autotests.Command())

	return command
}

var command = &cobra.Command{
	Use:     "environments",
	Short:   "Manage environments",
	Aliases: []string{"env"},
}
