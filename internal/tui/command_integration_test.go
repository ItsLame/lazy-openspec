package tui

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TestRunProcessLive drives the streaming command runner against the real
// openspec CLI (via `openspec --version`) and verifies it streams output and
// terminates with a done message. Skips when the CLI is absent.
func TestRunProcessLive(t *testing.T) {
	if _, err := exec.LookPath("openspec"); err != nil {
		t.Skip("openspec not installed")
	}
	ch := make(chan tea.Msg, 128)
	cmd := runProcess("openspec", []string{"--version"}, "version", ch)

	var lines []string
	deadline := time.After(15 * time.Second)
	msg := cmd() // first buffered message
	for {
		switch v := msg.(type) {
		case logLineMsg:
			lines = append(lines, v.line)
		case cmdDoneMsg:
			if v.err != nil {
				t.Fatalf("command failed: %v", v.err)
			}
			if len(lines) == 0 {
				t.Fatalf("expected streamed output, got none")
			}
			if !strings.ContainsAny(strings.Join(lines, ""), "0123456789") {
				t.Errorf("version output looks wrong: %v", lines)
			}
			return
		}
		select {
		case <-deadline:
			t.Fatalf("timed out; lines so far: %v", lines)
		default:
		}
		msg = waitForMsg(ch)()
	}
}
