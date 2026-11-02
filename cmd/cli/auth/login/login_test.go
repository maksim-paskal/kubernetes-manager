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
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newLoginCommand() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("address", "", "")
	cmd.Flags().String("login", "", "")
	cmd.Flags().String("password", "", "")

	return cmd
}

func TestSaveSettingsDoesNotWriteEnvironmentValues(t *testing.T) {
	// the environment variable is read by viper, but it must not end up in the file
	t.Setenv("KUBERNETES_MANAGER_AUTH_PASSWORD", "secret-from-env")

	viper.Reset()
	viper.SetEnvPrefix("KUBERNETES_MANAGER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	configFile := filepath.Join(t.TempDir(), ".kubernetes-manager.yaml")

	if err := saveSettings(newLoginCommand(), configFile); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(content), "secret-from-env") {
		t.Fatalf("environment password is written to the config file:\n%s", content)
	}
}

func TestSaveSettingsWritesGivenFlagsAndKeepsFileKeys(t *testing.T) {
	t.Parallel()

	configFile := filepath.Join(t.TempDir(), ".kubernetes-manager.yaml")

	if err := os.WriteFile(configFile, []byte("dev_user: alice\nauth:\n  address: https://old.example.com\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := newLoginCommand()
	if err := cmd.Flags().Set("password", "token"); err != nil {
		t.Fatal(err)
	}

	if err := saveSettings(cmd, configFile); err != nil {
		t.Fatal(err)
	}

	file := viper.New()
	file.SetConfigFile(configFile)

	if err := file.ReadInConfig(); err != nil {
		t.Fatal(err)
	}

	if got := file.GetString("auth.password"); got != "token" {
		t.Errorf("auth.password = %q, want %q", got, "token")
	}

	if got := file.GetString("auth.address"); got != "https://old.example.com" {
		t.Errorf("auth.address = %q, want the address kept from the file", got)
	}

	if got := file.GetString("dev_user"); got != "alice" {
		t.Errorf("dev_user = %q, want the key kept from the file", got)
	}

	info, err := os.Stat(configFile)
	if err != nil {
		t.Fatal(err)
	}

	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("config permissions = %o, want 600", perm)
	}
}
