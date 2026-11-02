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
package commands

import (
	"strings"

	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/global"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Command builds a new command on every call, so the tests can add it to their own roots.
func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "commands",
		Short: "List all commands and flags as JSON",
		Long:  "List all commands and flags as JSON. It is built from the command definitions, so it always matches the commands and does not depend on the help text.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root := cmd.Root()

			// always JSON, so the output is the same in a terminal and in a pipe
			global.OutputJSON(cmd, catalog{
				Name:        root.Name(),
				GlobalFlags: collectFlags(root.PersistentFlags()),
				Commands:    collectCommands(root, root, nil),
			})

			return nil
		},
	}
}

type catalog struct {
	Name        string        `json:"name"`
	GlobalFlags []flagInfo    `json:"globalFlags"`
	Commands    []commandInfo `json:"commands"`
}

type commandInfo struct {
	Path    string     `json:"path"`
	Use     string     `json:"use"`
	Short   string     `json:"short"`
	Long    string     `json:"long,omitempty"`
	Example string     `json:"example,omitempty"`
	Aliases []string   `json:"aliases,omitempty"`
	Flags   []flagInfo `json:"flags,omitempty"`
}

type flagInfo struct {
	Name      string `json:"name"`
	Shorthand string `json:"shorthand,omitempty"`
	Type      string `json:"type"`
	Default   string `json:"default,omitempty"`
	Usage     string `json:"usage"`
	Required  bool   `json:"required,omitempty"`
}

// collectCommands returns the runnable commands below parent, groups only lead to their children.
func collectCommands(root, parent *cobra.Command, items []commandInfo) []commandInfo {
	for _, sub := range parent.Commands() {
		// hidden, deprecated and help commands are not part of the interface
		if !sub.IsAvailableCommand() {
			continue
		}

		if sub.Runnable() {
			items = append(items, commandInfo{
				Path:    strings.TrimPrefix(sub.CommandPath(), root.Name()+" "),
				Use:     sub.Use,
				Short:   sub.Short,
				Long:    sub.Long,
				Example: sub.Example,
				Aliases: sub.Aliases,
				// inherited flags are listed once, at the root
				Flags: collectFlags(sub.LocalNonPersistentFlags()),
			})
		}

		items = collectCommands(root, sub, items)
	}

	return items
}

func collectFlags(set *pflag.FlagSet) []flagInfo {
	var items []flagInfo

	set.VisitAll(func(flag *pflag.Flag) {
		// cobra adds the help flag to every command
		if flag.Name == "help" {
			return
		}

		// MarkFlagRequired sets this annotation, the command fails without the flag
		_, required := flag.Annotations[cobra.BashCompOneRequiredFlag]

		items = append(items, flagInfo{
			Name:      flag.Name,
			Shorthand: flag.Shorthand,
			Type:      flag.Value.Type(),
			Default:   defaultValue(flag.DefValue),
			Usage:     flag.Usage,
			Required:  required,
		})
	})

	return items
}

// defaultValue hides the zero values, the default of a bool is false and the default of a list is empty.
func defaultValue(value string) string {
	switch value {
	case "", "false", "[]":
		return ""
	default:
		return value
	}
}
