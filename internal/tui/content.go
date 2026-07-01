package tui

import (
	"fmt"
	"strings"

	"github.com/itslame/lazy-openspec/internal/openspec"
	"github.com/itslame/lazy-openspec/internal/tasks"
)

// refreshMain rebuilds the main viewport content for the current state.
func (m Model) refreshMain() Model {
	if !m.ready {
		return m
	}
	d := m.layout()
	m.vp.Width, m.vp.Height = d.vpW, d.vpH
	m.md.SetWidth(d.vpW)

	switch m.screen {
	case screenChangeDetail:
		content, cursorLine := m.changeContent(d.vpW)
		m.vp.SetContent(content)
		if m.tab == tabTasks {
			m.ensureVisible(cursorLine, d.vpH)
		}
	case screenSpecDetail:
		m.vp.SetContent(m.specContent(d.vpW))
	default:
		m.vp.SetContent(m.dashboardPreview(d.vpW))
	}
	return m
}

// ensureVisible scrolls the viewport so line is within view.
func (m *Model) ensureVisible(line, height int) {
	off := m.vp.YOffset
	if line < off {
		m.vp.SetYOffset(line)
	} else if line >= off+height {
		m.vp.SetYOffset(line - height + 1)
	}
}

// dashboardPreview renders the right-pane preview for the dashboard.
func (m Model) dashboardPreview(width int) string {
	if m.loadErr != nil {
		return m.errorBlock()
	}
	if c, ok := m.selectedChange(); ok {
		return m.changeSummaryBlock(c, width)
	}
	if s, ok := m.selectedSpec(); ok {
		return fmt.Sprintf("%s\n%s\n\n%s",
			titleFocused.Render(s.Name),
			mutedText.Render(fmt.Sprintf("%d requirement(s)", s.RequirementCount)),
			mutedText.Render("Press enter to view requirements."))
	}
	return mutedText.Render("Nothing selected.")
}

// changeSummaryBlock renders a change's status + artifact checklist.
func (m Model) changeSummaryBlock(c openspec.ChangeSummary, width int) string {
	var b strings.Builder
	b.WriteString(titleFocused.Render(c.Name))
	b.WriteString("\n")
	b.WriteString(lifecycleLine(c))
	b.WriteString("\n\n")

	st, ok := m.statusCache[c.Name]
	if !ok {
		b.WriteString(mutedText.Render("Loading status…"))
		return b.String()
	}
	b.WriteString(mutedText.Render("Schema: " + st.SchemaName))
	b.WriteString("\n\n")
	b.WriteString(lipglossBold("Artifacts"))
	b.WriteString("\n")
	for _, a := range st.Artifacts {
		glyph := mutedText.Render("○")
		switch a.Status {
		case "done":
			glyph = glyphDone
		case "ready":
			glyph = lipglossColor("214", "◐")
		case "blocked":
			glyph = mutedText.Render("·")
		}
		b.WriteString(fmt.Sprintf("  %s %s\n", glyph, a.ID))
	}
	b.WriteString("\n")
	b.WriteString(mutedText.Render("Press enter to open artifacts."))
	return b.String()
}

// changeContent renders the body of the current artifact tab. It returns the
// content and, for the tasks tab, the line index of the selected task.
func (m Model) changeContent(width int) (string, int) {
	switch m.tab {
	case tabProposal, tabDesign:
		content, ok := m.detailCache[cacheKey(m.curChange, m.tab)]
		if !ok {
			return mutedText.Render("Loading…"), 0
		}
		if strings.TrimSpace(content) == "" {
			return mutedText.Render("No " + tabNames[m.tab] + " for this change yet."), 0
		}
		return m.md.Render(content), 0
	case tabTasks:
		content, ok := m.detailCache[cacheKey(m.curChange, tabTasks)]
		if !ok {
			return mutedText.Render("Loading…"), 0
		}
		return m.tasksTab(content, width)
	case tabSpecs:
		d, ok := m.changeDetail[m.curChange]
		if !ok {
			return mutedText.Render("Loading…"), 0
		}
		return m.sem.ChangeSpecs(d.Deltas, width), 0
	}
	return "", 0
}

// tasksTab renders the grouped checklist with the selected task marked, and
// returns the line index of that task so the viewport can keep it in view.
func (m Model) tasksTab(content string, width int) (string, int) {
	groups := tasks.Parse(content)
	if len(groups) == 0 {
		return mutedText.Render("No tasks defined yet."), 0
	}
	var lines []string
	cursorLine := 0
	idx := 0
	for gi, g := range groups {
		if gi > 0 {
			lines = append(lines, "")
		}
		title := g.Title
		if g.Number != "" {
			title = g.Number + ". " + title
		}
		bar := m.sem.ProgressBar(g.Completed(), g.Total(), 10)
		lines = append(lines, fmt.Sprintf("%s  %s %d/%d",
			lipglossBold(title), bar, g.Completed(), g.Total()))
		for _, t := range g.Tasks {
			pointer := "  "
			if idx == m.taskCursor {
				pointer = lipglossColor("39", "▸ ")
				cursorLine = len(lines)
			}
			glyph := mutedText.Render("☐")
			text := t.Number + " " + t.Text
			if t.Done {
				glyph = glyphDone
				text = faint(text)
			}
			lines = append(lines, pointer+glyph+" "+text)
			idx++
		}
	}
	if m.curArchived {
		lines = append(lines, "", mutedText.Render("(archived — read-only)"))
	}
	return strings.Join(lines, "\n"), cursorLine
}

// specContent renders a spec's requirements and records per-requirement line
// offsets for n/p navigation.
func (m *Model) specContent(width int) string {
	if m.specDetail == nil {
		return mutedText.Render("Loading spec…")
	}
	reqs := m.specDetail.Requirements
	if len(reqs) == 0 {
		return mutedText.Render("No requirements found for this spec.")
	}
	var lines []string
	m.reqOffsets = m.reqOffsets[:0]
	for i, r := range reqs {
		if i > 0 {
			lines = append(lines, "", "")
		}
		m.reqOffsets = append(m.reqOffsets, len(lines))
		block := m.sem.Requirement(i+1, r.Text, r.Scenarios, width)
		lines = append(lines, strings.Split(block, "\n")...)
	}
	return strings.Join(lines, "\n")
}

// errorBlock renders a friendly message for load failures.
func (m Model) errorBlock() string {
	switch {
	case m.loadErr == openspec.ErrCLINotFound:
		return errText.Render("openspec CLI not found") + "\n\n" +
			mutedText.Render("Install it first:\n  npm install -g @fission-ai/openspec")
	case m.loadErr == openspec.ErrNoRoot:
		return errText.Render("No OpenSpec project found") + "\n\n" +
			mutedText.Render("Run `openspec init` in your project, or launch lazy-openspec\ninside a directory that has an openspec/ root.")
	default:
		return errText.Render("Failed to load") + "\n\n" + mutedText.Render(m.loadErr.Error())
	}
}

func lifecycleLine(c openspec.ChangeSummary) string {
	switch c.Lifecycle() {
	case openspec.Draft:
		return glyphDraft + " " + mutedText.Render("draft")
	case openspec.Completed:
		return glyphDone + " " + mutedText.Render("completed")
	default:
		return fmt.Sprintf("%s %s  %d/%d tasks", glyphActive, mutedText.Render("active"),
			c.CompletedTasks, c.TotalTasks)
	}
}
