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
package root

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"charm.land/fang/v2"
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/auth"
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/clusters"
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/commands"
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/environments"
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/global"
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/profiles"
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/remote_servers"
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/services"
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	// cancelCommand stops the deadline of --timeout when the command is finished.
	cancelCommand context.CancelFunc
	rootCmd       = &cobra.Command{
		Use:   global.CLIName,
		Short: "kubernetes manager CLI tool",
		Long:  `kubernetes manager CLI tool for managing environments`,
	}
)

// Execute executes the root command.
func Execute(ctx context.Context) error {
	registerCommands()

	return fang.Execute(ctx,
		rootCmd,
		fang.WithoutCompletions(),
		fang.WithoutManpage(),
	)
}

func registerCommands() {
	cobra.OnInitialize(initConfig)
	cobra.OnFinalize(func() {
		if cancelCommand != nil {
			cancelCommand()
		}
	})

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/."+global.CLIName+".yaml)")
	rootCmd.PersistentFlags().Bool("debug", false, "debug logs")
	rootCmd.PersistentFlags().Bool("json", false, "output in JSON format")
	rootCmd.PersistentFlags().Bool("ids", false, "print only the first column of each table, e.g. environment ids")
	rootCmd.PersistentFlags().Duration("timeout", 0, "limit the whole command, e.g. 30s or 2m, 0 is no limit")

	// the two output modes are exclusive, otherwise one of them would be silently ignored
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		idsOutput, _ := cmd.Flags().GetBool("ids")

		if jsonOutput && idsOutput {
			return errors.New("--json and --ids cannot be used together")
		}

		timeout, err := cmd.Flags().GetDuration("timeout")
		if err != nil {
			return fmt.Errorf("invalid --timeout: %w", err)
		}

		if timeout > 0 {
			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			cancelCommand = cancel

			cmd.SetContext(ctx)
		}

		return nil
	}

	rootCmd.AddCommand(version.Command())
	rootCmd.AddCommand(auth.Command())
	rootCmd.AddCommand(remote_servers.Command())
	rootCmd.AddCommand(environments.Command())
	rootCmd.AddCommand(profiles.Command())
	rootCmd.AddCommand(services.Command())
	rootCmd.AddCommand(clusters.Command())
	rootCmd.AddCommand(commands.Command())

	showAliases(rootCmd)
}

// showAliases appends the aliases to the short description, fang lists only the usage and the short description.
func showAliases(cmd *cobra.Command) {
	if len(cmd.Aliases) > 0 {
		cmd.Short += " (alias: " + strings.Join(cmd.Aliases, ", ") + ")"
	}

	for _, sub := range cmd.Commands() {
		showAliases(sub)
	}
}

func initConfig() {
	// a config in the current folder is used only when it exists, a new config (login) is always created in the home directory
	localConfig := "." + global.CLIName + ".yaml"
	localInfo, localErr := os.Stat(localConfig)

	switch {
	case cfgFile != "":
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	case localErr == nil && !localInfo.IsDir():
		viper.SetConfigFile(localConfig)
	default:
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName("." + global.CLIName)
	}

	// nested keys map to variables, e.g. auth.password -> KUBERNETES_MANAGER_AUTH_PASSWORD
	viper.SetEnvPrefix(global.EnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err == nil {
		slog.Debug("Using config file", "file", viper.ConfigFileUsed())

		return
	}

	// a missing file is fine, login creates it and other commands report the missing settings
	var notFound viper.ConfigFileNotFoundError
	if errors.As(err, &notFound) || os.IsNotExist(err) {
		return
	}

	cobra.CheckErr(fmt.Errorf("failed to read config file %s: %w", viper.ConfigFileUsed(), err))
}
