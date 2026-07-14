package tui

import (
	"fmt"
	"strings"

	"github.com/itslame/lazy-openspec/internal/openspec"
	"github.com/itslame/lazy-openspec/internal/tasks"
)

// refreshMain rebuilds the main viewport content for the current selection. The
// previewed item follows the nav selection regardless of which pane is focused.
func (m Model) refreshMain() Model {
	if !m.ready {
		return m
	}
	d := m.layout()
	m.vp.Width, m.vp.Height = d.vpW, d.vpH
	m.md.SetWidth(d.vpW)

	var content string
	taskCursorLine := -1
	if m.loadErr != nil {
		content = m.errorBlock()
	} else if _, ok := m.selectedChange(); ok {
		var cl int
		content, cl = m.changeContent(d.vpW)
		if m.tab == tabTasks {
			taskCursorLine = cl
		}
	} else if _, ok := m.selectedSpec(); ok {
		content = m.specContent(d.vpW)
	} else {
		content = mutedText.Render("Nothing selected.")
	}

	// Incremental search: highlight matches and record their line offsets.
	if m.search.query != "" {
		content, m.search.matches = highlightMatches(content, m.search.query, m.search.idx)
		if m.search.idx >= len(m.search.matches) {
			m.search.idx = 0
		}
	} else {
		m.search.matches = nil
	}

	m.vp.SetContent(content)

	// A search match takes scroll precedence; otherwise keep the task cursor visible.
	if len(m.search.matches) > 0 {
		m.vp.SetYOffset(m.search.matches[m.search.idx])
	} else if taskCursorLine >= 0 {
		m.ensureVisible(taskCursorLine, d.vpH)
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

// overviewTab renders the overview tab: a change's lifecycle + artifact
// checklist. Archived changes are sourced from disk; active ones from status.
func (m Model) overviewTab(c openspec.ChangeSummary, width int) string {
	if m.curArchived {
		return m.archivedOverviewBlock(c)
	}
	return m.changeSummaryBlock(c, width)
}

// changeSummaryBlock renders an active change's status + artifact checklist.
func (m Model) changeSummaryBlock(c openspec.ChangeSummary, width int) string {
	var b strings.Builder
	b.WriteString(titleFocused.Render(c.Name))
	b.WriteString("\n")
	b.WriteString(lifecycleLine(c))
	b.WriteString("\n\n")

	st, ok := m.statusCache[c.Name]
	if !ok {
		if err, failed := m.statusErr[c.Name]; failed {
			b.WriteString(errText.Render("Failed to load status"))
			b.WriteString("\n")
			b.WriteString(mutedText.Render(err.Error()))
			return b.String()
		}
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
			glyph = lipglossColor("3", "◐")
		case "blocked":
			glyph = mutedText.Render("·")
		}
		b.WriteString(fmt.Sprintf("  %s %s\n", glyph, a.ID))
	}
	b.WriteString("\n")
	b.WriteString(mutedText.Render("enter: focus preview · [ ]: switch tabs"))
	return b.String()
}

// archivedOverviewBlock renders an archived change's overview from disk (task
// counts and artifact presence), avoiding the CLI which cannot resolve it.
func (m Model) archivedOverviewBlock(c openspec.ChangeSummary) string {
	var b strings.Builder
	b.WriteString(titleFocused.Render(c.Name))
	b.WriteString("  ")
	b.WriteString(faint("(archived)"))
	b.WriteString("\n\n")

	ov, ok := m.archivedOv[c.Name]
	if !ok {
		b.WriteString(mutedText.Render("Loading…"))
		return b.String()
	}
	if ov.total > 0 {
		b.WriteString(fmt.Sprintf("%s %s  %d/%d tasks", glyphDone, mutedText.Render("archived"), ov.completed, ov.total))
	} else {
		b.WriteString(glyphDone + " " + mutedText.Render("archived"))
	}
	b.WriteString("\n\n")
	b.WriteString(lipglossBold("Artifacts"))
	b.WriteString("\n")
	artRow := func(present bool, name string) {
		glyph := mutedText.Render("○")
		if present {
			glyph = glyphDone
		}
		b.WriteString(fmt.Sprintf("  %s %s\n", glyph, name))
	}
	artRow(ov.hasProposal, "proposal")
	b.WriteString(fmt.Sprintf("  %s %s\n", mutedText.Render("·"), mutedText.Render("specs (merged into main)")))
	artRow(ov.hasDesign, "design")
	artRow(ov.hasTasks, "tasks")
	b.WriteString("\n")
	b.WriteString(mutedText.Render("enter: focus preview · [ ]: switch tabs · read-only"))
	return b.String()
}

// changeContent renders the body of the current artifact tab. It returns the
// content and, for the tasks tab, the line index of the selected task.
func (m Model) changeContent(width int) (string, int) {
	switch m.tab {
	case tabOverview:
		c, ok := m.selectedChange()
		if !ok {
			return mutedText.Render("Nothing selected."), 0
		}
		return m.overviewTab(c, width), 0
	case tabProposal, tabDesign:
		key := cacheKey(m.curChange, m.tab)
		if msg, failed := m.detailErr[key]; failed {
			return errText.Render("Failed to load "+tabNames[m.tab]) + "\n\n" + mutedText.Render(msg), 0
		}
		content, ok := m.detailCache[key]
		if !ok {
			return mutedText.Render("Loading…"), 0
		}
		if strings.TrimSpace(content) == "" {
			return mutedText.Render("No " + tabNames[m.tab] + " for this change yet."), 0
		}
		return m.md.Render(content), 0
	case tabTasks:
		key := cacheKey(m.curChange, tabTasks)
		if msg, failed := m.detailErr[key]; failed {
			return errText.Render("Failed to load tasks") + "\n\n" + mutedText.Render(msg), 0
		}
		content, ok := m.detailCache[key]
		if !ok {
			return mutedText.Render("Loading…"), 0
		}
		return m.tasksTab(content, width)
	case tabSpecs:
		if m.curArchived {
			return mutedText.Render("Spec deltas were merged into the main specs when this change was archived."), 0
		}
		if err, failed := m.specsErr[m.curChange]; failed {
			return errText.Render("Failed to load specs") + "\n\n" + mutedText.Render(err.Error()), 0
		}
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
				pointer = lipglossColor("6", "▸ ")
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
	if err, failed := m.specErr[m.curSpec]; failed {
		return errText.Render("Failed to load spec") + "\n\n" + mutedText.Render(err.Error())
	}
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
