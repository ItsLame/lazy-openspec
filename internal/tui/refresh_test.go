package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/itslame/lazy-openspec/internal/openspec"
)

// feedCmd applies a message and returns the updated model *and* the command it
// produced, so tests can assert whether a refresh was dispatched.
func feedCmd(m Model, msg tea.Msg) (Model, tea.Cmd) {
	tm, cmd := m.Update(msg)
	return tm.(Model), cmd
}

// runCmd executes a command (flattening batches) and returns the messages it
// produced.
func runCmd(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	switch t := cmd().(type) {
	case nil:
		return nil
	case tea.BatchMsg:
		var out []tea.Msg
		for _, c := range t {
			out = append(out, runCmd(c)...)
		}
		return out
	default:
		return []tea.Msg{t}
	}
}

// onProposalOf selects a change, switches the preview to its proposal tab, and
// seeds that artifact's cache with body.
func onProposalOf(m Model, change, body string) Model {
	for m.curChange != change {
		m = feed(m, key("j"))
	}
	if m.tab != tabProposal {
		m = feed(m, key("]")) // overview -> proposal
	}
	return feed(m, artifactMsg{change: change, tab: tabProposal, content: body})
}

// TestRefreshReloadsPreviewedArtifact pins the bug this change fixes: a refresh
// must re-dispatch the loader for the artifact already on screen, not just
// re-run the list queries, so the preview stops rendering pre-refresh content.
func TestRefreshReloadsPreviewedArtifact(t *testing.T) {
	m := onProposalOf(seeded(), "add-x", "STALEBODY")
	if !strings.Contains(m.View(), "STALEBODY") {
		t.Fatalf("expected the cached proposal in the preview:\n%s", m.View())
	}

	// Pressing `r` must dispatch a loader for the *visible* artifact even though it
	// is already cached (ensurePreviewLoaded skips it as a cache hit). This drives
	// the real key path, so it also guards refreshAll's wiring — not just
	// reloadPreview in isolation. The list loaders in the same batch shell out to
	// the openspec CLI; their result is irrelevant here (an error message is fine),
	// only the artifact loader's dispatch is asserted.
	_, cmd := feedCmd(m, key("r"))
	found := false
	for _, msg := range runCmd(cmd) {
		if am, ok := msg.(artifactMsg); ok && am.change == "add-x" && am.tab == tabProposal {
			found = true
		}
	}
	if !found {
		t.Fatalf("`r` must re-dispatch the previewed artifact's loader, or the preview stays stale")
	}

	// Fresh content replaces the cached copy.
	m = feed(m, artifactMsg{change: "add-x", tab: tabProposal, content: "FRESHBODY"})
	out := m.View()
	if strings.Contains(out, "STALEBODY") || !strings.Contains(out, "FRESHBODY") {
		t.Fatalf("refreshed content should replace the cached version:\n%s", out)
	}
}

// TestRefreshDropsStaleCachesKeepsVisible: `r` invalidates the per-item caches
// (not just the lists), but carries the on-screen entries over so the preview
// does not flash "Loading…".
func TestRefreshDropsStaleCachesKeepsVisible(t *testing.T) {
	m := onProposalOf(seeded(), "add-x", "VISIBLEBODY")
	// A cached artifact and status for a change that is *not* on screen.
	m = feed(m, artifactMsg{change: "draft-y", tab: tabProposal, content: "OTHERBODY"})
	m = feed(m, statusMsg{change: "draft-y", st: openspec.Status{SchemaName: "spec-driven"}})

	m = feed(m, key("r"))

	if _, ok := m.detailCache[cacheKey("add-x", tabProposal)]; !ok {
		t.Errorf("the visible artifact must be carried over, or the preview flashes Loading…")
	}
	if _, ok := m.statusCache["add-x"]; !ok {
		t.Errorf("the visible change's status must be carried over")
	}
	if _, ok := m.detailCache[cacheKey("draft-y", tabProposal)]; ok {
		t.Errorf("off-screen artifact caches must be dropped so they reload fresh")
	}
	if _, ok := m.statusCache["draft-y"]; ok {
		t.Errorf("off-screen status caches must be dropped so they reload fresh")
	}
	if !strings.Contains(m.View(), "VISIBLEBODY") {
		t.Errorf("the preview should keep its content until fresh data lands:\n%s", m.View())
	}
}

// TestRefreshPreservesSelectionAcrossReorder: a reload that reorders the list
// keeps the same change selected, rather than leaving the index pointing at a
// different row.
func TestRefreshPreservesSelectionAcrossReorder(t *testing.T) {
	m := seeded() // add-x (active, idx 0), draft-y (draft, idx 1)
	m = feed(m, key("j"))
	if m.curChange != "draft-y" {
		t.Fatalf("setup: expected draft-y selected, got %q", m.curChange)
	}

	// Reload where add-x completed and draft-y became active: sortChanges now puts
	// draft-y first, so the old index 1 would land on add-x.
	m = feed(m, changesMsg{list: openspec.ChangeList{
		Changes: []openspec.ChangeSummary{
			{Name: "add-x", CompletedTasks: 5, TotalTasks: 5},   // completed -> sorts last
			{Name: "draft-y", CompletedTasks: 1, TotalTasks: 3}, // active -> sorts first
		},
		Root: openspec.Root{Path: "/tmp"},
	}})

	if got := m.selectedName(); got != "draft-y" {
		t.Fatalf("refresh must keep the same change selected across a reorder, got %q", got)
	}
	if m.sel[panelChanges] != 0 {
		t.Fatalf("selection should follow draft-y to its new index 0, got %d", m.sel[panelChanges])
	}
	if m.curChange != "draft-y" {
		t.Fatalf("the preview should still follow draft-y, got %q", m.curChange)
	}
}

// TestRefreshVanishedSelectionFallsBack: when the selected change is gone after a
// reload (e.g. archived externally), the selection lands on a valid row without
// erroring.
func TestRefreshVanishedSelectionFallsBack(t *testing.T) {
	m := seeded()
	m = feed(m, key("j")) // select draft-y
	m = feed(m, changesMsg{list: openspec.ChangeList{
		Changes: []openspec.ChangeSummary{{Name: "add-x", CompletedTasks: 2, TotalTasks: 5}},
		Root:    openspec.Root{Path: "/tmp"},
	}})

	c, ok := m.selectedChange()
	if !ok {
		t.Fatalf("selection should fall back to a valid row when its item vanishes")
	}
	if c.Name != "add-x" {
		t.Fatalf("expected the surviving change selected, got %q", c.Name)
	}
	if strings.Contains(m.View(), "draft-y") {
		t.Errorf("the vanished change should be gone from the view:\n%s", m.View())
	}
}

// TestRefreshKeepsScrollOffset: refreshed content re-renders at the same scroll
// position, and clamps (without panicking) when the content shrinks.
func TestRefreshKeepsScrollOffset(t *testing.T) {
	var long strings.Builder
	for i := 0; i < 200; i++ {
		fmt.Fprintf(&long, "- proposal line %d\n", i)
	}
	m := onProposalOf(seeded(), "add-x", long.String())
	m = feed(m, tea.KeyMsg{Type: tea.KeyEnter}) // focus preview
	m = feed(m, key("G"))                       // scroll to the bottom

	off := m.vp.YOffset
	if off == 0 {
		t.Fatalf("setup: expected the long proposal to be scrollable")
	}

	m = feed(m, key("r")) // refresh: content is carried over, so nothing moves
	if m.vp.YOffset != off {
		t.Errorf("refresh moved the scroll offset: %d -> %d", off, m.vp.YOffset)
	}
	// Fresh (identical) content lands: still no movement.
	m = feed(m, artifactMsg{change: "add-x", tab: tabProposal, content: long.String()})
	if m.vp.YOffset != off {
		t.Errorf("re-rendering identical content moved the scroll offset: %d -> %d", off, m.vp.YOffset)
	}
	// The artifact shrank on disk: the offset must clamp, not dangle past the end.
	m = feed(m, artifactMsg{change: "add-x", tab: tabProposal, content: "- only line\n"})
	if m.vp.YOffset > off {
		t.Errorf("scroll offset should clamp when content shrinks, got %d", m.vp.YOffset)
	}
}

// TestFocusRefreshTriggersAndDebounces covers the focus-driven refresh guards:
// a bare focus does nothing, blur→focus refreshes, and a second cycle inside the
// debounce window does not.
func TestFocusRefreshTriggersAndDebounces(t *testing.T) {
	m := seeded()
	// A cached entry for an off-screen change: dropped iff a refresh actually runs.
	stale := func(m Model) Model {
		return feed(m, statusMsg{change: "draft-y", st: openspec.Status{SchemaName: "spec-driven"}})
	}
	m = stale(m)

	// Focus without a preceding blur must not refresh (terminals emit one right
	// after focus reporting is enabled).
	m, cmd := feedCmd(m, tea.FocusMsg{})
	if cmd != nil {
		t.Fatalf("focus without a prior blur must not refresh")
	}
	if _, ok := m.statusCache["draft-y"]; !ok {
		t.Fatalf("focus without a prior blur must not drop caches")
	}

	// Blur, then focus: refresh.
	m = feed(m, tea.BlurMsg{})
	if !m.blurred {
		t.Fatalf("blur should mark the model blurred")
	}
	m, cmd = feedCmd(m, tea.FocusMsg{})
	if cmd == nil {
		t.Fatalf("regaining focus after a blur should refresh")
	}
	if m.blurred {
		t.Fatalf("focus should clear the blurred flag")
	}
	if _, ok := m.statusCache["draft-y"]; ok {
		t.Fatalf("the focus refresh should have invalidated the per-item caches")
	}

	// A second blur→focus inside the debounce window must not refresh again.
	m = stale(m)
	m = feed(m, tea.BlurMsg{})
	m, cmd = feedCmd(m, tea.FocusMsg{})
	if cmd != nil {
		t.Fatalf("focus flapping inside the debounce window must refresh at most once")
	}
	if _, ok := m.statusCache["draft-y"]; !ok {
		t.Fatalf("the debounced focus event must not have refreshed")
	}
}

// TestFocusRefreshSkippedWhileRunning: a streaming command's own completion
// refresh covers the update, so focus must not spawn a competing one.
func TestFocusRefreshSkippedWhileRunning(t *testing.T) {
	m := seeded()
	m.running = true
	m = feed(m, tea.BlurMsg{})
	m, cmd := feedCmd(m, tea.FocusMsg{})
	if cmd != nil {
		t.Fatalf("focus refresh must be skipped while a command is running")
	}
	if _, ok := m.statusCache["add-x"]; !ok {
		t.Fatalf("the skipped refresh must not have touched the caches")
	}
}
