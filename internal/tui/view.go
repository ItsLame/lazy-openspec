package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/itslame/lazy-openspec/internal/openspec"
)

// View renders the whole UI.
func (m Model) View() string {
	if m.quitting {
		return ""
	}
	if !m.ready {
		return "starting lazy-openspec…"
	}
	if m.width < minCols || m.height < minRows {
		return m.tooSmall()
	}
	if m.confirm != nil {
		return m.overlay("Confirm", m.confirm.prompt+"\n\n"+mutedText.Render("y = yes    n = no"))
	}
	if m.showHelp {
		return m.overlay("Keybindings", helpBody(m.helpEntries()))
	}
	if m.showActions {
		return m.overlay("Actions", helpBody(actionEntries))
	}

	d := m.layout()
	body := lipgloss.JoinHorizontal(lipgloss.Top, m.renderLeft(d), m.renderMain(d))
	return lipgloss.JoinVertical(lipgloss.Left, body, m.renderLog(d), m.renderHint())
}

func (m Model) tooSmall() string {
	msg := fmt.Sprintf("Terminal too small.\nNeed at least %d×%d, have %d×%d.",
		minCols, minRows, m.width, m.height)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, errText.Render(msg))
}

// ---- left column ------------------------------------------------------------

func (m Model) renderLeft(d dims) string {
	var boxes []string
	for p := panel(0); p < numPanels; p++ {
		h := d.panelH[p]
		// Only the two border lines are reserved: the title lives in the border.
		bodyH := h - 2
		if bodyH < 1 {
			bodyH = 1
		}
		body, focusLine := m.panelBody(p, d.leftW-4)
		body = windowLines(body, bodyH, focusLine)
		boxes = append(boxes, panelBox(m.panelTitle(p), body, d.leftW, h, m.focus == p))
	}
	return lipgloss.JoinVertical(lipgloss.Left, boxes...)
}

// panelTitle renders a panel's border title, appending the active filter query
// (e.g. `[2]─Specs /nav`) so a narrowed row set always has a visible cause.
// withBorderTitle truncates it if it outgrows the frame, so the border holds.
func (m Model) panelTitle(p panel) string {
	if !m.filter.appliesTo(p) {
		return panelTitles[p]
	}
	q := "/" + m.filter.query
	if m.filter.typing {
		q += "▏"
	}
	return panelTitles[p] + " " + q
}

// emptyListText picks a panel's empty state. A filter matching nothing says so
// explicitly rather than leaving an unexplained empty box.
func (m Model) emptyListText(p panel, width int, empty string) string {
	if m.filter.appliesTo(p) && m.filter.query != "" {
		return mutedText.Render(trunc("No matches for "+m.filter.query, width))
	}
	return mutedText.Render(trunc(empty, width))
}

func (m Model) panelBody(p panel, width int) (string, int) {
	switch p {
	case panelChanges:
		return m.changesList(width)
	case panelSpecs:
		return m.specsList(width)
	case panelArchive:
		return m.archiveList(width)
	}
	return "", 0
}

func (m Model) changesList(width int) (string, int) {
	vis := m.visibleChanges()
	if len(vis) == 0 {
		return m.emptyListText(panelChanges, width, "No changes yet."), 0
	}
	var lines []string
	focusLine := 0
	last := openspec.Lifecycle(-1)
	// Iterating the visible rows means a lifecycle group whose every row is
	// filtered out never emits a header.
	for i, c := range vis {
		if lc := c.Lifecycle(); lc != last {
			if len(lines) > 0 {
				lines = append(lines, "")
			}
			lines = append(lines, faint(groupName(lc)))
			last = lc
		}
		sel := i == m.sel[panelChanges]
		if sel {
			focusLine = len(lines)
		}
		lines = append(lines, changeRow(c, width, sel && m.focus == panelChanges))
	}
	return strings.Join(lines, "\n"), focusLine
}

// specsList renders each spec as a right-aligned requirement-count gutter
// followed by the name, so the counts form a readable column down the left edge
// (the way lazygit leads a commit row with its timestamp) instead of trailing
// names of varying length.
func (m Model) specsList(width int) (string, int) {
	vis := m.visibleSpecs()
	if len(vis) == 0 {
		return m.emptyListText(panelSpecs, width, "No specs yet."), 0
	}
	// The gutter is as wide as the widest *visible* count, so a filter that hides
	// the only two-digit spec reclaims the column instead of leaving it ragged.
	counts := make([]string, len(vis))
	gutter := 0
	for i, s := range vis {
		counts[i] = fmt.Sprintf("%dr", s.RequirementCount)
		if n := len([]rune(counts[i])); n > gutter {
			gutter = n
		}
	}
	// Floor the name's budget: with the count leading the row, a truncation bug
	// here would render a bare count with an empty name.
	nameW := width - gutter - 2
	if nameW < 3 {
		nameW = 3
	}
	var lines []string
	focusLine := 0
	for i, s := range vis {
		sel := i == m.sel[panelSpecs]
		if sel {
			focusLine = len(lines)
		}
		count := fmt.Sprintf("%*s", gutter, counts[i])
		name := trunc(s.Name, nameW)
		if sel && m.focus == panelSpecs {
			lines = append(lines, selectedItem.Render(fit(count+"  "+name, width)))
		} else {
			lines = append(lines, faint(count)+"  "+name)
		}
	}
	return strings.Join(lines, "\n"), focusLine
}

func (m Model) archiveList(width int) (string, int) {
	vis := m.visibleArchived()
	if len(vis) == 0 {
		return m.emptyListText(panelArchive, width, "No archived changes."), 0
	}
	var lines []string
	focusLine := 0
	for i, c := range vis {
		sel := i == m.sel[panelArchive]
		if sel {
			focusLine = len(lines)
		}
		if sel && m.focus == panelArchive {
			lines = append(lines, selectedItem.Render(fit("▫ "+c.Name, width)))
		} else {
			// Overhead is "▫ " (2); truncate the name to fit.
			lines = append(lines, faint("▫ "+trunc(c.Name, width-2)))
		}
	}
	return strings.Join(lines, "\n"), focusLine
}

func changeRow(c openspec.ChangeSummary, width int, selected bool) string {
	glyph := "○"
	switch c.Lifecycle() {
	case openspec.Active:
		glyph = "◉"
	case openspec.Completed:
		glyph = "✓"
	}
	suffix := ""
	if c.Lifecycle() == openspec.Active {
		suffix = fmt.Sprintf(" %d%%", c.Percent())
	}
	if selected {
		return selectedItem.Render(fit(glyph+" "+c.Name+suffix, width))
	}
	cg := glyph
	switch c.Lifecycle() {
	case openspec.Active:
		cg = lipglossColor("3", glyph)
	case openspec.Completed:
		cg = lipglossColor("2", glyph)
	default:
		cg = mutedText.Render(glyph)
	}
	// Overhead is the glyph (1) + a space (1) + the suffix; truncate the name to fit.
	name := trunc(c.Name, width-2-len([]rune(suffix)))
	return cg + " " + name + faint(suffix)
}

func groupName(lc openspec.Lifecycle) string {
	switch lc {
	case openspec.Active:
		return "Active"
	case openspec.Draft:
		return "Draft"
	default:
		return "Completed"
	}
}

// ---- main pane --------------------------------------------------------------

func (m Model) renderMain(d dims) string {
	title, subtitle := m.mainTitles()
	inner := strings.Join([]string{subtitle, m.vp.View(), m.scrollIndicator()}, "\n")
	focused := m.activePane == panePreview
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor(focused)).
		Padding(0, 1).
		Width(d.mainW - 2).
		Height(d.bodyH - 2).
		MaxWidth(d.mainW).
		MaxHeight(d.bodyH).
		Render(inner)
	return withBorderTitle(box, title, focused)
}

func (m Model) mainTitles() (string, string) {
	ts := borderTitleStyle(m.activePane == panePreview)
	if c, ok := m.selectedChange(); ok {
		title := ts.Render(c.Name)
		if m.focus == panelArchive {
			title += " " + faint("(archived)")
		}
		return title, m.tabBar()
	}
	if s, ok := m.selectedSpec(); ok {
		sub := ""
		if m.specDetail != nil && len(m.specDetail.Requirements) > 0 {
			sub = mutedText.Render(fmt.Sprintf("requirement %d/%d", m.reqIdx+1, len(m.specDetail.Requirements)))
		}
		return ts.Render("spec: " + s.Name), sub
	}
	return ts.Render("Details"), mutedText.Render("preview")
}

func (m Model) tabBar() string {
	parts := make([]string, 0, numTabs)
	for t := artifactTab(0); t < numTabs; t++ {
		if t == m.tab {
			parts = append(parts, tabActive.Render(tabNames[t]))
		} else {
			parts = append(parts, tabInactive.Render(tabNames[t]))
		}
	}
	return strings.Join(parts, faint(" · "))
}

func (m Model) scrollIndicator() string {
	pct := int(m.vp.ScrollPercent() * 100)
	if pct < 0 {
		pct = 0
	}
	s := faint(fmt.Sprintf("── scroll %d%%", pct))
	if m.search.active() {
		s += "  " + m.searchStatus()
	}
	return s
}

// searchStatus renders the search prompt and match counter for the scroll line.
func (m Model) searchStatus() string {
	prompt := hintKey.Render("/") + m.search.query
	if m.search.typing {
		prompt += "▏"
	}
	count := ""
	switch {
	case m.search.query == "":
	case len(m.search.matches) == 0:
		count = faint("  no matches")
	default:
		count = faint(fmt.Sprintf("  match %d/%d", m.search.idx+1, len(m.search.matches)))
	}
	return prompt + count
}

// ---- bottom: log + hint -----------------------------------------------------

func (m Model) renderLog(d dims) string {
	inner := d.logH - 2
	if inner < 1 {
		inner = 1
	}
	var shown []string
	if len(m.logs) > inner {
		shown = m.logs[len(m.logs)-inner:]
	} else {
		shown = append([]string{}, m.logs...)
		for len(shown) < inner {
			shown = append(shown, "")
		}
	}
	title := logTitle.Render("Command log")
	if m.running {
		title += " " + lipglossColor("3", "● running")
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colNone).
		Padding(0, 1).
		Width(m.width - 2).
		Height(d.logH - 2).
		MaxWidth(m.width).
		MaxHeight(d.logH).
		Render(strings.Join(shown, "\n"))
	return withBorderTitle(box, title, false)
}

func (m Model) renderHint() string {
	var b strings.Builder
	for i, bind := range m.shortHints() {
		if i > 0 {
			b.WriteString(faint("  "))
		}
		b.WriteString(hintKey.Render(bind.keys) + " " + hintDesc.Render(bind.desc))
	}
	return lipgloss.NewStyle().Width(m.width).MaxWidth(m.width).Render(b.String())
}

// shortHints returns the hint-bar bindings for the active pane. In the preview
// the bindings depend on whether a change (tabs) or spec (requirements) is shown.
// From the list, [ ] steer the preview's sections and / filters the panel.
func (m Model) shortHints() []binding {
	if m.activePane == panePreview {
		if _, ok := m.selectedSpec(); ok {
			return []binding{{"↑↓ j/k", "scroll"}, {"[ ]", "req"}, {"/", "search"}, {"n/N", "match"}, {"esc", "list"}, {"?", "help"}}
		}
		return []binding{{"↑↓ j/k", "scroll"}, {"[ ]", "tab"}, {"space", "toggle"}, {"/", "search"}, {"esc", "list"}, {"?", "help"}}
	}
	if m.filter.active() {
		return []binding{{"esc", "clear filter"}, {"⏎", "confirm"}, {"↑↓", "move"}, {"?", "help"}}
	}
	return []binding{{"tab", "panel"}, {"↑↓", "move"}, {"[ ]", "section"}, {"/", "filter"}, {"⏎", "focus"}, {"x", "actions"}, {"?", "help"}, {"q", "quit"}}
}

// ---- overlays ---------------------------------------------------------------

var actionEntries = []binding{
	{"v", "validate change"},
	{"a", "show apply instructions"},
	{"A", "archive change"},
	{"esc", "close"},
}

func (m Model) overlay(title, body string) string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colFocus).
		Padding(1, 3).
		Render(body)
	box = withBorderTitle(box, borderTitleStyle(true).Render(title), true)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func helpBody(entries []binding) string {
	w := 0
	for _, e := range entries {
		if len(e.keys) > w {
			w = len(e.keys)
		}
	}
	var lines []string
	for _, e := range entries {
		lines = append(lines, hintKey.Render(fmt.Sprintf("%-*s", w, e.keys))+"  "+hintDesc.Render(e.desc))
	}
	return strings.Join(lines, "\n")
}

// ---- text helpers -----------------------------------------------------------

// fit truncates or right-pads plain text to exactly width columns.
func fit(s string, width int) string {
	if width < 0 {
		width = 0
	}
	r := []rune(s)
	if len(r) > width {
		if width <= 1 {
			return string(r[:width])
		}
		return string(r[:width-1]) + "…"
	}
	return s + strings.Repeat(" ", width-len(r))
}

// trunc shortens plain text to at most width columns, adding an ellipsis when it
// overflows. Unlike fit it does not right-pad, so uncolored rows keep their
// natural length while still never exceeding the panel width.
func trunc(s string, width int) string {
	if width < 0 {
		width = 0
	}
	r := []rune(s)
	if len(r) <= width {
		return s
	}
	if width <= 1 {
		return string(r[:width])
	}
	return string(r[:width-1]) + "…"
}

// windowLines returns a slice of s's lines that fits height, keeping the focus
// line visible, padding with blanks when there are fewer lines than height.
func windowLines(s string, height, focus int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= height {
		for len(lines) < height {
			lines = append(lines, "")
		}
		return strings.Join(lines, "\n")
	}
	start := focus - height/2
	if start < 0 {
		start = 0
	}
	if start+height > len(lines) {
		start = len(lines) - height
	}
	return strings.Join(lines[start:start+height], "\n")
}
