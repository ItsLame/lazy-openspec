package render

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/itslame/lazy-openspec/internal/openspec"
	"github.com/itslame/lazy-openspec/internal/tasks"
)

// scenarioKeyword matches "- **WHEN** ..." style scenario lines (also THEN, AND,
// GIVEN, WHERE), case-insensitively.
var scenarioKeyword = regexp.MustCompile(`(?i)^\s*-\s*\*\*(WHEN|THEN|AND|GIVEN|WHERE)\*\*\s*(.*)$`)

// inlineCode strips markdown inline-code backticks; the surrounding terminal
// styling already provides enough contrast.
var inlineCode = regexp.MustCompile("`([^`]*)`")

// Semantic renders OpenSpec's structured blocks with dedicated styling. Colors
// degrade automatically on NO_COLOR terminals (via lipgloss/termenv), and the
// glyph + layout structure keeps output readable without color.
type Semantic struct {
	when    lipgloss.Style
	then    lipgloss.Style
	other   lipgloss.Style
	req     lipgloss.Style
	badge   map[string]lipgloss.Style
	group   lipgloss.Style
	checked lipgloss.Style
	pending lipgloss.Style
	dim     lipgloss.Style
	filled  lipgloss.Style
	empty   lipgloss.Style
}

// NewSemantic builds the semantic renderer with its style palette. Colors are
// drawn from the terminal's ANSI 16-color palette (indices 0-15) so the main
// pane matches the panel chrome and the user's terminal theme, and degrade
// automatically on NO_COLOR terminals via lipgloss/termenv.
func NewSemantic() *Semantic {
	bold := lipgloss.NewStyle().Bold(true)
	return &Semantic{
		when:  bold.Foreground(lipgloss.Color("2")), // green
		then:  bold.Foreground(lipgloss.Color("6")), // cyan
		other: bold.Foreground(lipgloss.Color("3")), // yellow
		req:   bold.Underline(true),
		badge: map[string]lipgloss.Style{
			"ADDED":    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2")), // green
			"MODIFIED": lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("3")), // yellow
			"REMOVED":  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1")), // red
			"RENAMED":  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6")), // cyan
		},
		group:   bold,
		checked: lipgloss.NewStyle().Foreground(lipgloss.Color("2")), // green
		pending: lipgloss.NewStyle().Faint(true),
		dim:     lipgloss.NewStyle().Faint(true),
		filled:  lipgloss.NewStyle().Foreground(lipgloss.Color("2")), // green
		empty:   lipgloss.NewStyle().Faint(true),
	}
}

// clean strips inline-code backticks for plain readable text.
func clean(s string) string { return inlineCode.ReplaceAllString(s, "$1") }

// wrap word-wraps body text to width, used for hanging-indent layouts.
func wrap(s string, width int) string {
	if width < 4 {
		width = 4
	}
	return lipgloss.NewStyle().Width(width).Render(s)
}

// Scenario renders a raw scenario block (its rawText) as aligned WHEN/THEN rows
// with a hanging indent so wrapped lines line up under the text, not the keyword.
func (s *Semantic) Scenario(raw string, width int) string {
	var rows []string
	for _, line := range strings.Split(raw, "\n") {
		m := scenarioKeyword.FindStringSubmatch(line)
		if m == nil {
			if t := strings.TrimSpace(line); t != "" {
				rows = append(rows, wrap(clean(t), width))
			}
			continue
		}
		kw := strings.ToUpper(m[1])
		style := s.other
		switch kw {
		case "WHEN", "GIVEN":
			style = s.when
		case "THEN":
			style = s.then
		}
		label := style.Render(fmt.Sprintf("%-5s", kw))
		body := wrap(clean(strings.TrimSpace(m[2])), width-8)
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, "  "+label+"  ", body))
	}
	return strings.Join(rows, "\n")
}

// Requirement renders a requirement heading, its text, and its scenarios.
func (s *Semantic) Requirement(index int, text string, scenarios []openspec.Scenario, width int) string {
	var b strings.Builder
	b.WriteString(s.req.Render(fmt.Sprintf("Requirement %d", index)))
	b.WriteString("\n")
	b.WriteString(wrap(clean(text), width))
	for _, sc := range scenarios {
		b.WriteString("\n\n")
		b.WriteString(s.Scenario(sc.RawText, width))
	}
	return b.String()
}

// ChangeSpecs renders the specs tab of a change: its delta operations grouped by
// spec, each with a colored operation badge and its requirements/scenarios.
func (s *Semantic) ChangeSpecs(deltas []openspec.Delta, width int) string {
	if len(deltas) == 0 {
		return s.dim.Render("No spec deltas in this change.")
	}
	var blocks []string
	for _, d := range deltas {
		badge := s.badge[strings.ToUpper(d.Operation)]
		header := badge.Render(strings.ToUpper(d.Operation)) + "  " + lipgloss.NewStyle().Bold(true).Render(d.Spec)
		reqs := d.Requirements
		var body []string
		for i, r := range reqs {
			body = append(body, s.Requirement(i+1, r.Text, r.Scenarios, width-2))
		}
		blocks = append(blocks, header+"\n"+indent(strings.Join(body, "\n\n"), 2))
	}
	return strings.Join(blocks, "\n\n")
}

// SpecRequirements renders a standalone spec's requirements.
func (s *Semantic) SpecRequirements(reqs []openspec.Requirement, width int) string {
	if len(reqs) == 0 {
		return s.dim.Render("No requirements found for this spec.")
	}
	var blocks []string
	for i, r := range reqs {
		blocks = append(blocks, s.Requirement(i+1, r.Text, r.Scenarios, width))
	}
	return strings.Join(blocks, "\n\n")
}

// Tasks renders tasks.md as grouped checklists with per-group progress and
// checkbox glyphs; completed tasks are dimmed.
func (s *Semantic) Tasks(groups []tasks.Group, width int) string {
	if len(groups) == 0 {
		return s.dim.Render("No tasks defined yet.")
	}
	var blocks []string
	for _, g := range groups {
		title := g.Title
		if g.Number != "" {
			title = g.Number + ". " + title
		}
		head := s.group.Render(title)
		bar := s.ProgressBar(g.Completed(), g.Total(), 10)
		header := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(maxInt(1, width-18)).Render(head),
			fmt.Sprintf("  %s %d/%d", bar, g.Completed(), g.Total()),
		)
		var items []string
		for _, t := range g.Tasks {
			glyph := s.pending.Render("☐")
			text := t.Number + " " + t.Text
			if t.Done {
				glyph = s.checked.Render("✔")
				text = s.dim.Render(text)
			}
			items = append(items, "  "+glyph+" "+text)
		}
		blocks = append(blocks, header+"\n"+strings.Join(items, "\n"))
	}
	return strings.Join(blocks, "\n\n")
}

// ProgressBar renders a fixed-width bar of filled/empty cells.
func (s *Semantic) ProgressBar(done, total, cells int) string {
	if cells < 1 {
		cells = 1
	}
	fill := 0
	if total > 0 {
		fill = int(float64(done) / float64(total) * float64(cells))
	}
	if fill > cells {
		fill = cells
	}
	return s.filled.Render(strings.Repeat("█", fill)) + s.empty.Render(strings.Repeat("░", cells-fill))
}

// indent prefixes every line with n spaces.
func indent(s string, n int) string {
	pad := strings.Repeat(" ", n)
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = pad + lines[i]
	}
	return strings.Join(lines, "\n")
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
