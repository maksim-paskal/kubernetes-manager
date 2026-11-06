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
package kubeconfig

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/global"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	command.Flags().String("path", "", "Path to save the kubeconfig file")
	command.Flags().Bool("force", false, "Replace the file at --path if it already exists")

	return command
}

// writeFile saves the kubeconfig readable only by the owner, an existing file is replaced only with force.
func writeFile(path, content string, force bool) error {
	flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if force {
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}

	file, err := os.OpenFile(path, flags, 0o600)
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			return fmt.Errorf("%s already exists, use --force to replace it", path)
		}

		return err
	}

	// the mode of an existing file is not changed by OpenFile, set it explicitly
	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()

		return err
	}

	if _, err := file.WriteString(content); err != nil {
		_ = file.Close()

		return err
	}

	return file.Close()
}

var command = &cobra.Command{
	Use:   "kubeconfig <id>",
	Short: "Get kubeconfig for environment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := global.Client(cmd)
		if err != nil {
			return err
		}

		kubeconfig, err := client.GetKubeConfig(cmd.Context(), args[0])
		if err != nil {
			return err
		}

		path, _ := cmd.Flags().GetString("path")
		if path == "" {
			global.Output(cmd, kubeconfig)

			return nil
		}

		force, _ := cmd.Flags().GetBool("force")

		return writeFile(path, kubeconfig, force)
	},
}
