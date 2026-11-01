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
package login

import (
	"os"
	"path/filepath"

	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/global"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Command() *cobra.Command {
	command.Flags().String("address", "", "Address of the kubernetes manager")
	command.Flags().String("login", "", "Login for basic authentication")
	command.Flags().String("password", "", "Password for basic authentication")
	cobra.CheckErr(viper.BindPFlag("auth.address", command.Flags().Lookup("address")))
	cobra.CheckErr(viper.BindPFlag("auth.login", command.Flags().Lookup("login")))
	cobra.CheckErr(viper.BindPFlag("auth.password", command.Flags().Lookup("password")))

	return command
}

var command = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the kubernetes manager",
	RunE: func(cmd *cobra.Command, _ []string) error {
		client, err := global.Client(cmd)
		if err != nil {
			return errors.Wrap(err, "failed to create client")
		}

		_, err = client.GetUser(cmd.Context())
		if err != nil {
			return errors.Wrap(err, "failed to authenticate with the kubernetes manager")
		}

		// config contains the password in plain text, create it readable only by the owner
		viper.SetConfigPermissions(0o600)

		// an explicit --config file (or the existing default config) is overwritten,
		// SafeWriteConfig only creates the default config and ignores the --config path
		var writeErr error
		if viper.ConfigFileUsed() != "" {
			writeErr = viper.WriteConfig()
		} else {
			writeErr = viper.SafeWriteConfig()
		}

		if writeErr != nil {
			return errors.Wrap(writeErr, "failed to write config")
		}

		// config contains the password in plain text, restrict access to the owner
		configFile := viper.ConfigFileUsed()
		if configFile == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return errors.Wrap(err, "failed to get home directory")
			}

			configFile = filepath.Join(home, "."+global.CLIName+".yaml")
		}

		if err := os.Chmod(configFile, 0o600); err != nil {
			return errors.Wrap(err, "failed to restrict config permissions")
		}

		global.Output(cmd, "OK")

		return nil
	},
}
