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
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// AddYesFlag adds --yes to a command that destroys something.
func AddYesFlag(cmd *cobra.Command) {
	cmd.Flags().BoolP("yes", "y", false, "do not ask for confirmation")
}

// Confirm asks the question on a terminal. Without a terminal it refuses unless --yes is set,
// so scripts and agents never destroy anything by accident.
func Confirm(cmd *cobra.Command, question string) error {
	if yes, _ := cmd.Flags().GetBool("yes"); yes {
		return nil
	}

	stdin, ok := cmd.InOrStdin().(*os.File)
	if !ok || !term.IsTerminal(int(stdin.Fd())) {
		return errors.New("confirmation required, use --yes to run without asking")
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "%s [y/N]: ", question)

	answer, _ := bufio.NewReader(stdin).ReadString('\n')
	if answer = strings.ToLower(strings.TrimSpace(answer)); answer != "y" && answer != "yes" {
		return errors.New("aborted")
	}

	return nil
}
