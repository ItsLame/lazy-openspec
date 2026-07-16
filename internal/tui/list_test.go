package tui

import (
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/itslame/lazy-openspec/internal/openspec"
)

// threeSpecs seeds the Specs panel with names that exercise both the filter
// (two contain "nav") and the count gutter (one- and two-digit counts).
func threeSpecs(m Model) Model {
	return feed(m, specsMsg{list: openspec.SpecList{Specs: []openspec.SpecSummary{
		{Name: "auth", RequirementCount: 4},
		{Name: "change-navigation", RequirementCount: 12},
		{Name: "spec-navigation", RequirementCount: 3},
	}}})
}

// TestStickyArtifactTabAcrossSelection pins the sticky-tab behaviour: picking an
// artifact tab once and moving the selection previews *that* artifact for the
// newly selected change, rather than snapping back to the overview.
func TestStickyArtifactTabAcrossSelection(t *testing.T) {
	m := seeded()
	m = feed(m, key("]")) // list-focused: overview -> proposal
	if m.tab != tabProposal {
		t.Fatalf("] should advance the tab to proposal, got %v", m.tab)
	}
	m = feed(m, key("j")) // move to the next change (draft-y)
	if m.tab != tabProposal {
		t.Fatalf("the active tab must persist across selection moves, got %v", m.tab)
	}
	if m.curChange != "draft-y" {
		t.Fatalf("selection should have moved to draft-y, got %q", m.curChange)
	}
	// The preview shows the newly selected change's proposal, not the old one's.
	m = feed(m, artifactMsg{change: "draft-y", tab: tabProposal, content: "DRAFTYPROPOSAL"})
	if !strings.Contains(m.View(), "DRAFTYPROPOSAL") {
		t.Fatalf("expected draft-y's proposal in the preview:\n%s", m.View())
	}
}

// TestListSectionKeysSteerPreview: [ / ] move between the preview's sections —
// a change's artifact tabs and a spec's requirements — without taking focus off
// the list.
func TestListSectionKeysSteerPreview(t *testing.T) {
	m := seeded()
	m = feed(m, key("]"))
	if m.tab != tabProposal {
		t.Fatalf("] on a list-focused change should advance the tab, got %v", m.tab)
	}
	if m.activePane != paneNav {
		t.Fatalf("] must leave focus on the list, got pane %v", m.activePane)
	}
	m = feed(m, key("[")) // back to overview
	if m.tab != tabOverview {
		t.Fatalf("[ should step back to overview, got %v", m.tab)
	}

	// A spec has no tabs, so [ / ] step through its requirements instead. Enough
	// requirements that the content overflows the viewport and can actually scroll.
	reqs := make([]openspec.Requirement, 12)
	for i := range reqs {
		reqs[i] = openspec.Requirement{
			Text:      "The system SHALL satisfy requirement number " + string(rune('a'+i)),
			Scenarios: []openspec.Scenario{{RawText: "- **WHEN** a thing happens\n- **THEN** another follows"}},
		}
	}
	m = feed(m, key("2")) // focus Specs
	m = feed(m, specDetailMsg{id: "auth", detail: openspec.SpecDetail{
		ID: "auth", Name: "auth", Requirements: reqs,
	}})
	if len(m.reqOffsets) != len(reqs) {
		t.Fatalf("expected %d requirement offsets, got %d", len(reqs), len(m.reqOffsets))
	}
	before := m.vp.YOffset
	m = feed(m, key("]"))
	if m.reqIdx != 1 {
		t.Fatalf("] on a list-focused spec should advance reqIdx to 1, got %d", m.reqIdx)
	}
	if m.vp.YOffset == before {
		t.Fatalf("] should scroll the preview to the next requirement (offset stayed %d)", before)
	}
	if m.activePane != paneNav {
		t.Fatalf("] must leave focus on the list, got pane %v", m.activePane)
	}
}

// TestFilterQueryCapturesGlobalKeys is the highest-value guard in this change: a
// filter query is free text typed into a handler whose fallthrough contains q
// (quit) and A (archive). A missing or mis-ordered guard is data-destructive.
func TestFilterQueryCapturesGlobalKeys(t *testing.T) {
	m := seeded()
	m = feed(m, key("/"))
	if !m.filter.typing {
		t.Fatalf("/ on the list should open the filter")
	}
	for _, k := range []string{"q", "r", "v", "x", "A"} {
		m = feed(m, key(k))
	}
	if m.filter.query != "qrvxA" {
		t.Fatalf("global keys should be typed into the query, got %q", m.filter.query)
	}
	if m.quitting {
		t.Fatalf("typing q must not quit")
	}
	if m.showActions {
		t.Fatalf("typing x must not open the actions overlay")
	}
	if m.confirm != nil {
		t.Fatalf("typing A must not open the archive confirmation")
	}
	if m.running {
		t.Fatalf("typing r/v must not run a command")
	}

	// Once the query is confirmed it no longer captures keystrokes.
	m = feed(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.filter.typing {
		t.Fatalf("enter should confirm the filter")
	}
	if m.filter.query != "qrvxA" {
		t.Fatalf("enter should keep the query, got %q", m.filter.query)
	}
	m = feed(m, key("q"))
	if !m.quitting {
		t.Fatalf("q should quit again once the filter is confirmed")
	}
}

// TestFilterNarrowsAndSelectionFollows: the previewed item always matches the
// highlighted row, j steps between matching rows only, and esc restores the full
// list with the same item still selected.
func TestFilterNarrowsAndSelectionFollows(t *testing.T) {
	m := threeSpecs(seeded())
	m = feed(m, key("2")) // focus Specs (auth selected)
	m = feed(m, key("/"))
	m = feed(m, key("nav"))

	vis := m.visibleSpecs()
	if len(vis) != 2 || vis[0].Name != "change-navigation" || vis[1].Name != "spec-navigation" {
		t.Fatalf("filter should narrow to the two *nav* specs, got %+v", vis)
	}
	// "auth" was selected but no longer matches, so the selection falls back to
	// the first visible row and the preview follows it.
	if m.curSpec != "change-navigation" {
		t.Fatalf("preview should follow the highlighted row, got %q", m.curSpec)
	}
	if got := m.selectedName(); got != "change-navigation" {
		t.Fatalf("highlighted row should be change-navigation, got %q", got)
	}
	// The query is visible on the panel, so the narrowed list has a cause.
	if !strings.Contains(ansi.Strip(m.View()), "/nav") {
		t.Fatalf("the active query should be shown on the panel:\n%s", m.View())
	}

	m = feed(m, tea.KeyMsg{Type: tea.KeyEnter}) // confirm; list keys work again
	m = feed(m, key("j"))
	if m.selectedName() != "spec-navigation" || m.curSpec != "spec-navigation" {
		t.Fatalf("j should step to the next matching row, got %q (preview %q)", m.selectedName(), m.curSpec)
	}
	m = feed(m, key("j")) // at the last match: must not step onto a hidden row
	if m.curSpec != "spec-navigation" {
		t.Fatalf("j must not land on a filtered-out row, got %q", m.curSpec)
	}

	m = feed(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.filter.active() {
		t.Fatalf("esc should clear the filter")
	}
	if len(m.visibleSpecs()) != 3 {
		t.Fatalf("esc should restore the full list, got %d rows", len(m.visibleSpecs()))
	}
	if m.selectedName() != "spec-navigation" {
		t.Fatalf("esc should leave the same item selected, got %q", m.selectedName())
	}
}

// TestFilterIsScopedAndClearedOnPanelSwitch: the filter narrows only the panel
// that opened it, and moving focus away clears it so no panel stays silently
// narrowed.
func TestFilterIsScopedAndClearedOnPanelSwitch(t *testing.T) {
	m := threeSpecs(seeded())
	m = feed(m, key("2"))
	m = feed(m, key("/"))
	m = feed(m, key("nav"))
	if len(m.visibleSpecs()) != 2 {
		t.Fatalf("Specs should be narrowed, got %d rows", len(m.visibleSpecs()))
	}
	if len(m.visibleChanges()) != 2 {
		t.Fatalf("the Changes panel must not be filtered, got %d rows", len(m.visibleChanges()))
	}
	m = feed(m, tea.KeyMsg{Type: tea.KeyTab}) // focus another panel
	if m.filter.active() {
		t.Fatalf("switching panels should clear the filter")
	}
	if len(m.visibleSpecs()) != 3 {
		t.Fatalf("the Specs panel should be restored to its full list, got %d rows", len(m.visibleSpecs()))
	}
}

// TestFilterNoMatches: a query matching nothing says so, and the preview drops to
// its empty state rather than keeping stale content.
func TestFilterNoMatches(t *testing.T) {
	m := seeded()
	m = feed(m, key("/"))
	m = feed(m, key("zzz"))
	if len(m.visibleChanges()) != 0 {
		t.Fatalf("expected no matching changes, got %d", len(m.visibleChanges()))
	}
	if _, ok := m.selectedChange(); ok {
		t.Fatalf("nothing should be selected when no row matches")
	}
	out := ansi.Strip(m.View())
	if !strings.Contains(out, "No matches") {
		t.Fatalf("expected an explicit no-matches message:\n%s", out)
	}
	if !strings.Contains(out, "Nothing selected") {
		t.Fatalf("expected the preview's empty state, not stale content:\n%s", out)
	}
}

// TestSpecCountGutterAligns: counts of differing width right-align to a common
// column so every name starts at the same column, a long name is ellipsised
// without breaking the panel, and no row renders as a bare count.
func TestSpecCountGutterAligns(t *testing.T) {
	m := seeded()
	m = feed(m, specsMsg{list: openspec.SpecList{Specs: []openspec.SpecSummary{
		{Name: "artifact-rendering", RequirementCount: 3},
		{Name: "change-operations", RequirementCount: 12},
		{Name: "a-really-long-spec-capability-name-that-overflows-the-panel", RequirementCount: 7},
	}}})
	m = feed(m, key("2")) // focus Specs, so row 0 also exercises the selected path

	const width = 30
	body, _ := m.specsList(width)
	lines := strings.Split(body, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 spec rows, got %d", len(lines))
	}

	row := regexp.MustCompile(`^(\s*\d+r)  (\S.*?)\s*$`)
	gutter := -1
	for i, ln := range lines {
		if w := lipgloss.Width(ln); w > width {
			t.Errorf("row %d is %d wide, exceeding the %d-column panel: %q", i, w, width, ln)
		}
		mt := row.FindStringSubmatch(ansi.Strip(ln))
		if mt == nil {
			t.Fatalf("row %d is not '<count>  <name>': %q", i, ansi.Strip(ln))
		}
		if gutter == -1 {
			gutter = len(mt[1])
		} else if len(mt[1]) != gutter {
			t.Errorf("row %d: count column is %d wide, want %d — counts must right-align",
				i, len(mt[1]), gutter)
		}
		if strings.TrimSpace(mt[2]) == "" {
			t.Errorf("row %d rendered a bare count with an empty name: %q", i, ansi.Strip(ln))
		}
	}
	if gutter != 3 { // widest visible count is "12r"
		t.Errorf("gutter should be as wide as the widest count (3), got %d", gutter)
	}
	if !strings.Contains(ansi.Strip(lines[2]), "…") {
		t.Errorf("the overflowing name should be ellipsised: %q", ansi.Strip(lines[2]))
	}
}
