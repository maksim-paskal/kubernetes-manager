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
package version

import (
	"runtime/debug"

	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/global"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return command
}

var command = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of the kubernetes manager CLI tool",
	Run: func(cmd *cobra.Command, _ []string) {
		global.Output(cmd, "kubernetes manager CLI tool version "+buildVersion())
	},
}

// buildVersion returns the module version and the git commit recorded by the Go toolchain,
// the CLI is not built with the release ldflags used by the server.
func buildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	version := info.Main.Version
	if version == "" {
		version = "(devel)"
	}

	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" && setting.Value != "" {
			return version + " (" + setting.Value + ")"
		}
	}

	return version
}
