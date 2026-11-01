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
package hosts

import (
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/global"
	"github.com/spf13/cobra"
)

const (
	hostTypeDefault  = "default"
	hostTypeInternal = "internal"
)

type host struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

func Command() *cobra.Command {
	return command
}

var command = &cobra.Command{
	Use:   "hosts <id>",
	Short: "List environment hosts",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := global.Client(cmd)
		if err != nil {
			return err
		}

		environment, err := client.GetEnvironment(cmd.Context(), args[0])
		if err != nil {
			return err
		}

		hosts := make([]host, 0, len(environment.Hosts)+len(environment.HostsInternal))
		for _, hostURL := range environment.Hosts {
			hosts = append(hosts, host{URL: hostURL, Type: hostTypeDefault})
		}

		for _, hostURL := range environment.HostsInternal {
			hosts = append(hosts, host{URL: hostURL, Type: hostTypeInternal})
		}

		rows := make([][]string, len(hosts))
		for i, h := range hosts {
			rows[i] = []string{h.URL, h.Type}
		}

		global.OutputTable(cmd,
			hosts,
			[]string{"URL", "TYPE"},
			rows,
		)

		return nil
	},
}
