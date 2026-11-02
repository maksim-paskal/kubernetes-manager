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

		// an explicit --config file (or the existing default config) is overwritten,
		// a new config is created in the home directory
		configFile := viper.ConfigFileUsed()
		if configFile == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return errors.Wrap(err, "failed to get home directory")
			}

			configFile = filepath.Join(home, "."+global.CLIName+".yaml")
		}

		if err := saveSettings(cmd, configFile); err != nil {
			return err
		}

		global.Output(cmd, "OK")

		return nil
	},
}

// saveSettings writes the config file from its own settings, not from viper: viper also holds the environment
// variables, so the password from KUBERNETES_MANAGER_AUTH_PASSWORD would be written in plain text.
// Only the flags given to login are set, the other keys stay as they are in the file.
func saveSettings(cmd *cobra.Command, configFile string) error {
	file := viper.New()
	file.SetConfigFile(configFile)

	if err := file.ReadInConfig(); err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "failed to read config")
	}

	for _, name := range []string{"address", "login", "password"} {
		if cmd.Flags().Changed(name) {
			value, _ := cmd.Flags().GetString(name)
			file.Set("auth."+name, value)
		}
	}

	// config contains the password in plain text, create it readable only by the owner
	file.SetConfigPermissions(0o600)

	if err := file.WriteConfigAs(configFile); err != nil {
		return errors.Wrap(err, "failed to write config")
	}

	// the write does not change the permissions of an existing file, restrict access to the owner
	if err := os.Chmod(configFile, 0o600); err != nil {
		return errors.Wrap(err, "failed to restrict config permissions")
	}

	return nil
}
