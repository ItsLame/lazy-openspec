package tui

import (
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/itslame/lazy-openspec/internal/openspec"
)

// TestLoadArchivedOverviewReadsDisk verifies the archive overview is derived from
// on-disk files (task counts + artifact presence) with no CLI involvement.
func TestLoadArchivedOverviewReadsDisk(t *testing.T) {
	dir := t.TempDir()
	if err := writeFile(filepath.Join(dir, "tasks.md"), "## 1. Group\n\n- [x] 1.1 a\n- [ ] 1.2 b\n"); err != nil {
		t.Fatal(err)
	}
	if err := writeFile(filepath.Join(dir, "proposal.md"), "# proposal"); err != nil {
		t.Fatal(err)
	}
	// design.md intentionally absent.
	msg, ok := loadArchivedOverview(dir, "c")().(archivedOverviewMsg)
	if !ok {
		t.Fatalf("expected archivedOverviewMsg")
	}
	if msg.ov.completed != 1 || msg.ov.total != 2 {
		t.Fatalf("task counts: got %d/%d, want 1/2", msg.ov.completed, msg.ov.total)
	}
	if !msg.ov.hasProposal || !msg.ov.hasTasks || msg.ov.hasDesign {
		t.Fatalf("artifact presence wrong: %+v", msg.ov)
	}
}

// key builds a rune KeyMsg (mirrors how the terminal delivers letter keys).
func key(s string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }

// TestChangePreviewLiveOnSelection: a selected change renders its overview tab in
// the preview pane with no enter required, while the list keeps focus.
func TestChangePreviewLiveOnSelection(t *testing.T) {
	m := seeded()
	if m.activePane != paneNav {
		t.Fatalf("nav pane should hold focus initially")
	}
	out := m.View()
	for _, want := range []string{"overview", "Artifacts", "spec-driven"} {
		if !strings.Contains(out, want) {
			t.Fatalf("live change preview missing %q:\n%s", want, out)
		}
	}
}

// TestSpecPreviewLiveOnSelection: selecting a spec loads and renders its
// requirements live, without pressing enter and without leaving nav focus.
func TestSpecPreviewLiveOnSelection(t *testing.T) {
	m := seeded()
	m = feed(m, key("2")) // focus Specs -> live-loads spec detail (cmd discarded by feed)
	if !strings.Contains(m.View(), "Loading spec") {
		t.Fatalf("expected loading placeholder before detail arrives:\n%s", m.View())
	}
	m = feed(m, specDetailMsg{id: "auth", detail: openspec.SpecDetail{
		ID: "auth", Name: "auth",
		Requirements: []openspec.Requirement{{Text: "The system SHALL authenticate the user"}},
	}})
	if m.activePane != paneNav {
		t.Fatalf("preview should render without leaving nav focus")
	}
	if !strings.Contains(m.View(), "SHALL authenticate") {
		t.Fatalf("spec requirements should render live on selection:\n%s", m.View())
	}
}

// TestArchivedPreviewDoesNotHang pins the archive bug fix: an archived change's
// overview is sourced from disk and never sticks on the status loader, and the
// specs tab explains the merge instead of hanging.
func TestArchivedPreviewDoesNotHang(t *testing.T) {
	m := seeded()
	m = feed(m, archivedMsg{items: []openspec.ChangeSummary{{Name: "2026-01-01-old-thing"}}})
	m = feed(m, key("3")) // focus Archive
	m = feed(m, archivedOverviewMsg{change: "2026-01-01-old-thing", ov: archivedOverview{
		completed: 3, total: 3, hasProposal: true, hasDesign: true, hasTasks: true,
	}})
	out := m.View()
	if strings.Contains(out, "Loading status") {
		t.Fatalf("archived overview stuck on status loading:\n%s", out)
	}
	if !strings.Contains(out, "archived") || !strings.Contains(out, "3/3 tasks") {
		t.Fatalf("archived overview should render from disk:\n%s", out)
	}
	// Move to the specs tab in the focused preview; it must explain the merge.
	m = feed(m, tea.KeyMsg{Type: tea.KeyEnter}) // focus preview
	m = feed(m, key("]"))                        // overview -> proposal
	m = feed(m, key("]"))                        // proposal -> specs
	if !strings.Contains(m.View(), "merged into the main specs") {
		t.Fatalf("archived specs tab should explain the merge:\n%s", m.View())
	}
}

// TestEnterFocusesPreviewEscReturns: enter moves focus into the preview pane and
// esc (with no active search) returns focus to the list.
func TestEnterFocusesPreviewEscReturns(t *testing.T) {
	m := seeded()
	m = feed(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.activePane != panePreview {
		t.Fatalf("enter should focus the preview pane")
	}
	if !strings.Contains(m.View(), "search") { // preview hint bar advertises search
		t.Fatalf("preview hint bar expected:\n%s", m.View())
	}
	m = feed(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.activePane != paneNav {
		t.Fatalf("esc should return focus to the list")
	}
}

// TestPreviewSearchJumpAndCycle: search finds matches, confirms, cycles with
// wraparound, and clears on esc.
func TestPreviewSearchJumpAndCycle(t *testing.T) {
	m := seeded()
	m = feed(m, tea.KeyMsg{Type: tea.KeyEnter}) // focus preview (add-x overview)
	m = feed(m, key("/"))                        // open search
	m = feed(m, key("tasks"))                    // type query
	if len(m.search.matches) < 2 {
		t.Fatalf("expected >=2 matches for 'tasks', got %d", len(m.search.matches))
	}
	m = feed(m, tea.KeyMsg{Type: tea.KeyEnter}) // confirm
	if m.search.typing {
		t.Fatalf("enter should confirm (exit typing) the search")
	}
	n := len(m.search.matches)
	for i := 0; i < n; i++ {
		m = feed(m, key("n")) // cycle a full loop
	}
	if m.search.idx != 0 {
		t.Fatalf("cycling %d matches should wrap back to 0, got %d", n, m.search.idx)
	}
	m = feed(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.search.active() {
		t.Fatalf("esc should clear the active search")
	}
}
