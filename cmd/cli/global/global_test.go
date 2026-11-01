package global_test

import (
	"bytes"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/maksim-paskal/kubernetes-manager/cmd/cli/global"
	"github.com/spf13/cobra"
)

func TestOutputTableBorderAligned(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().Bool("ids", false, "")

	out := &bytes.Buffer{}
	cmd.SetOut(out)

	global.OutputTable(cmd, nil, []string{"ID", "DESCRIPTION"}, [][]string{
		{"1", "1️⃣ Paket"},
		{"2", "🏘️ Apollo"},
		{"3", "ℹ️ InfoDesk"},
		{"4", "👼 Admin"},
		{"5", "plain"},
	})

	lines := bytes.Split(bytes.TrimSpace(out.Bytes()), []byte("\n"))

	for _, line := range lines {
		if got, want := lipgloss.Width(string(line)), lipgloss.Width(string(lines[0])); got != want {
			t.Errorf("line %q has width %d, want %d", line, got, want)
		}

		// variation selectors are measured differently by terminals
		if bytes.ContainsRune(line, '️') {
			t.Errorf("line %q contains a variation selector", line)
		}
	}
}
