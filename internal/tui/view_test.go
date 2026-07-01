package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/itslame/lazy-openspec/internal/openspec"
)

// feed applies a message and returns the updated concrete Model.
func feed(m Model, msg tea.Msg) Model {
	tm, _ := m.Update(msg)
	return tm.(Model)
}

func seeded() Model {
	m := New(openspec.New())
	m = feed(m, tea.WindowSizeMsg{Width: 100, Height: 32})
	m = feed(m, changesMsg{list: openspec.ChangeList{
		Changes: []openspec.ChangeSummary{
			{Name: "add-x", CompletedTasks: 2, TotalTasks: 5, Status: "in-progress"},
			{Name: "draft-y", CompletedTasks: 0, TotalTasks: 0},
		},
		Root: openspec.Root{Path: "/tmp"},
	}})
	m = feed(m, specsMsg{list: openspec.SpecList{Specs: []openspec.SpecSummary{{Name: "auth", RequirementCount: 4}}}})
	m = feed(m, statusMsg{change: "add-x", st: openspec.Status{
		SchemaName: "spec-driven",
		Artifacts: []openspec.ArtifactStatus{
			{ID: "proposal", Status: "done"},
			{ID: "tasks", Status: "ready"},
		},
	}})
	return m
}

func TestDashboardRenders(t *testing.T) {
	m := seeded()
	out := m.View()
	for _, want := range []string{"Changes", "Specs", "Archive", "add-x", "auth", "Command log"} {
		if !strings.Contains(out, want) {
			t.Fatalf("dashboard missing %q in:\n%s", want, out)
		}
	}
}

func TestNavigateAndOpenDoesNotPanic(t *testing.T) {
	m := seeded()
	// Move down, switch panels, open, switch tabs, back.
	for _, k := range []string{"down", "2", "1", "enter", "]", "]", "]", "[", "esc", "?"} {
		m = feed(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)})
		_ = m.View()
	}
}

func TestTooSmall(t *testing.T) {
	m := New(openspec.New())
	m = feed(m, tea.WindowSizeMsg{Width: 20, Height: 8})
	if !strings.Contains(m.View(), "too small") {
		t.Fatalf("expected too-small message")
	}
}

func TestCLINotFoundMessage(t *testing.T) {
	m := New(openspec.New())
	m = feed(m, tea.WindowSizeMsg{Width: 100, Height: 32})
	m = feed(m, changesMsg{err: openspec.ErrCLINotFound})
	if !strings.Contains(m.View(), "openspec CLI not found") {
		t.Fatalf("expected CLI-not-found message in:\n%s", m.View())
	}
}

func TestConfirmOverlayOnArchive(t *testing.T) {
	m := seeded()
	m = feed(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("A")})
	if m.confirm == nil {
		t.Fatalf("archive should open a confirm prompt")
	}
	if !strings.Contains(m.View(), "Archive") {
		t.Fatalf("confirm overlay not shown")
	}
	// Cancel it.
	m = feed(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.confirm != nil {
		t.Fatalf("esc should cancel confirm")
	}
}
