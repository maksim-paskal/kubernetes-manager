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
package global

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"regexp"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/maksim-paskal/kubernetes-manager/internal/client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const CLIName = "kubernetes-manager"

// EnvPrefix is the prefix of environment variables that override the config, e.g. KUBERNETES_MANAGER_AUTH_PASSWORD.
const EnvPrefix = "KUBERNETES_MANAGER"

// envReferenceRe matches a whole value that is a reference to an environment variable, e.g. ${KM_PASSWORD}.
var envReferenceRe = regexp.MustCompile(`^\$\{([A-Za-z_][A-Za-z0-9_]*)\}$`)

// resolveValue returns the environment variable for a value that is exactly ${NAME},
// any other value (including passwords with $ in them) is used as is.
func resolveValue(value string) string {
	if match := envReferenceRe.FindStringSubmatch(value); match != nil {
		return os.Getenv(match[1])
	}

	return value
}

func Client(cmd *cobra.Command) (*client.Client, error) {
	if debug, err := cmd.Flags().GetBool("debug"); err == nil && debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	cli := client.NewClient().
		WithBaseURL(resolveValue(viper.GetString("auth.address"))).
		WithBasicAuth(
			resolveValue(viper.GetString("auth.login")),
			resolveValue(viper.GetString("auth.password")),
		).
		WithDevUser(viper.GetString("dev_user"))

	if err := cli.Validate(); err != nil {
		// a config file that was given but does not exist is the most likely cause
		if file := viper.ConfigFileUsed(); file != "" {
			if _, statErr := os.Stat(file); os.IsNotExist(statErr) {
				return nil, errors.Wrapf(err, "config file %s does not exist", file)
			}
		}

		return nil, errors.Wrap(err, "invalid client configuration")
	}

	return cli, nil
}

// ForEachID runs action for every id. All ids are validated first, so a bad id does not leave a partial run behind.
// A failing id does not stop the rest, every failure is printed and the command fails at the end.
func ForEachID(cmd *cobra.Command, ids []string, validate func(id string) error, action func(id string) error) error {
	if validate != nil {
		for _, id := range ids {
			if err := validate(id); err != nil {
				return err
			}
		}
	}

	failed := 0

	for _, id := range ids {
		if err := action(id); err != nil {
			failed++

			cmd.PrintErrln(errors.Wrap(err, id))
		}
	}

	if failed > 0 {
		return errors.Errorf("%d of %d failed", failed, len(ids))
	}

	return nil
}

func OutputTable(cmd *cobra.Command, data any, headers []string, rows [][]string) {
	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		Output(cmd, data)

		return
	}

	// the first column is the value the next command takes, e.g. the environment id or the profile name
	idsOutput, _ := cmd.Flags().GetBool("ids")
	if idsOutput {
		for i := range rows {
			Output(cmd, strings.TrimSpace(rows[i][0]))
		}

		return
	}

	if len(rows) == 0 {
		// stderr keeps stdout clean for scripts, an empty result is not an error
		cmd.PrintErrln("no results found")

		return
	}

	cells := make([][]string, len(rows))
	for i := range rows {
		cells[i] = make([]string, len(rows[i]))
		for j := range rows[i] {
			cells[i][j] = tableCell(rows[i][j])
		}
	}

	fmt.Fprintln(cmd.OutOrStdout(), table.New().
		Border(lipgloss.NormalBorder()).
		StyleFunc(func(_, _ int) lipgloss.Style {
			return lipgloss.NewStyle().Padding(0, 1)
		}).
		Headers(headers...).
		Rows(cells...).
		Render())
}

// variationSelectorReplacer drops the emoji presentation selector and the keycap mark. Terminals disagree
// with the width calculation for such sequences (e.g. "1️⃣", "🏘️"), which shifts the table border.
var variationSelectorReplacer = strings.NewReplacer("\uFE0F", "", "\u20E3", "")

// tableCell makes a cell safe to measure, so the table border stays aligned.
func tableCell(value string) string {
	return variationSelectorReplacer.Replace(value)
}

func Output(cmd *cobra.Command, message any) {
	slog.Debug("message type", "type", reflect.TypeOf(message))

	jsonOutput, _ := cmd.Flags().GetBool("json")

	if jsonOutput {
		OutputJSON(cmd, message)

		return
	}

	switch message.(type) {
	case string:
		fmt.Fprintln(cmd.OutOrStdout(), message)
	default:
		OutputJSON(cmd, message)
	}
}

// OutputJSON prints the message as indented JSON, the output flags are not checked.
func OutputJSON(cmd *cobra.Command, message any) {
	b, err := json.MarshalIndent(message, "", "  ")
	if err != nil {
		cmd.PrintErrln("failed to marshal output:", err)

		return
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(b))
}
