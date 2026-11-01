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
package commands_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/commands"
	"github.com/spf13/cobra"
)

type catalog struct {
	Name        string `json:"name"`
	GlobalFlags []struct {
		Name string `json:"name"`
	} `json:"globalFlags"`
	Commands []struct {
		Path    string   `json:"path"`
		Aliases []string `json:"aliases"`
		Flags   []struct {
			Name      string `json:"name"`
			Shorthand string `json:"shorthand"`
			Type      string `json:"type"`
			Default   string `json:"default"`
		} `json:"flags"`
	} `json:"commands"`
}

func newRoot() *cobra.Command {
	root := &cobra.Command{Use: "kubernetes-manager"}
	root.PersistentFlags().Bool("json", false, "output in JSON format")
	root.PersistentFlags().Bool("ids", false, "print only the first column")

	group := &cobra.Command{Use: "environments"}
	get := &cobra.Command{
		Use:     "get <id>",
		Aliases: []string{"info"},
		Run:     func(*cobra.Command, []string) {},
	}
	get.Flags().String("name", "", "name of the environment")
	get.Flags().BoolP("yes", "y", false, "do not ask")
	get.Flags().StringSlice("service", []string{}, "services")

	group.AddCommand(get)
	group.AddCommand(&cobra.Command{Use: "hidden", Hidden: true, Run: func(*cobra.Command, []string) {}})
	root.AddCommand(group, commands.Command())

	root.SetArgs([]string{"commands"})

	return root
}

func TestCommandsListsRunnableCommands(t *testing.T) {
	t.Parallel()

	root := newRoot()
	out := &bytes.Buffer{}
	root.SetOut(out)

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	var result catalog
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, out.String())
	}

	if result.Name != "kubernetes-manager" {
		t.Errorf("name = %q", result.Name)
	}

	globals := map[string]bool{}
	for _, flag := range result.GlobalFlags {
		globals[flag.Name] = true
	}

	if !globals["json"] || !globals["ids"] {
		t.Errorf("global flags = %v, want json and ids", globals)
	}

	paths := map[string]int{}

	for i, command := range result.Commands {
		paths[command.Path] = i
	}

	for _, path := range []string{"environments get", "commands"} {
		if _, ok := paths[path]; !ok {
			t.Errorf("command %q is missing, got %v", path, paths)
		}
	}

	for _, path := range []string{"environments", "environments hidden", "help"} {
		if _, ok := paths[path]; ok {
			t.Errorf("command %q must not be listed", path)
		}
	}

	get := result.Commands[paths["environments get"]]
	if len(get.Aliases) != 1 || get.Aliases[0] != "info" {
		t.Errorf("aliases = %v, want [info]", get.Aliases)
	}

	flags := map[string]struct{ shorthand, typ, def string }{}
	for _, flag := range get.Flags {
		flags[flag.Name] = struct{ shorthand, typ, def string }{flag.Shorthand, flag.Type, flag.Default}
	}

	if _, ok := flags["help"]; ok {
		t.Error("help flag must not be listed")
	}

	if _, ok := flags["json"]; ok {
		t.Error("inherited flag json must not be listed on the command")
	}

	if flags["yes"].shorthand != "y" || flags["yes"].typ != "bool" || flags["yes"].def != "" {
		t.Errorf("yes flag = %+v", flags["yes"])
	}

	if flags["service"].typ != "stringSlice" || flags["service"].def != "" {
		t.Errorf("service flag = %+v, empty list default must be hidden", flags["service"])
	}
}

func TestCommandsIsJSONWithOutputFlags(t *testing.T) {
	t.Parallel()

	root := newRoot()
	root.SetArgs([]string{"commands", "--ids"})

	out := &bytes.Buffer{}
	root.SetOut(out)

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	if !json.Valid(out.Bytes()) {
		t.Fatalf("output is not JSON with --ids:\n%s", out.String())
	}
}
