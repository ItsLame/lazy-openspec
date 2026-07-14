package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// searchState drives the incremental find over the focused preview. While
// typing is true the query line is being edited; matches holds the rendered
// line indices that contain the query and idx is the current match within it.
type searchState struct {
	typing  bool
	query   string
	matches []int
	idx     int
}

// active reports whether a search is in progress (being typed or applied).
func (s searchState) active() bool { return s.typing || s.query != "" }

// clearSearch resets all search state.
func (m *Model) clearSearch() { m.search = searchState{} }

var (
	// searchHL highlights every match; searchHLCur distinguishes the current one.
	searchHL    = lipgloss.NewStyle().Background(colActive).Foreground(lipgloss.Color("0"))
	searchHLCur = lipgloss.NewStyle().Background(colFocus).Foreground(lipgloss.Color("0")).Bold(true)
)

// asciiLower lowercases ASCII letters while preserving byte length, so indices
// computed on the folded copy line up with the original string. Non-ASCII runes
// are left as-is (they still match exactly, just not case-insensitively).
func asciiLower(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 'a' - 'A'
		}
	}
	return string(b)
}

// highlightMatches finds the lines of content (ANSI stripped) that contain
// query and returns the content with those lines re-emitted as plain text with
// each occurrence highlighted, plus the matching line indices in order. The
// current match (by line) is styled distinctly. Matching is case-insensitive.
func highlightMatches(content, query string, current int) (string, []int) {
	if query == "" {
		return content, nil
	}
	lines := strings.Split(content, "\n")
	lq := asciiLower(query)

	var matches []int
	for i, line := range lines {
		if strings.Contains(asciiLower(ansi.Strip(line)), lq) {
			matches = append(matches, i)
		}
	}
	if len(matches) == 0 {
		return content, nil
	}

	curLine := -1
	if current >= 0 && current < len(matches) {
		curLine = matches[current]
	}
	for _, i := range matches {
		style := searchHL
		if i == curLine {
			style = searchHLCur
		}
		lines[i] = highlightLine(ansi.Strip(lines[i]), query, style)
	}
	return strings.Join(lines, "\n"), matches
}

// highlightLine wraps every case-insensitive occurrence of query in plain with
// style, preserving the original text's casing.
func highlightLine(plain, query string, style lipgloss.Style) string {
	lp := asciiLower(plain)
	lq := asciiLower(query)
	var b strings.Builder
	i := 0
	for {
		j := strings.Index(lp[i:], lq)
		if j < 0 {
			b.WriteString(plain[i:])
			break
		}
		start := i + j
		end := start + len(query)
		b.WriteString(plain[i:start])
		b.WriteString(style.Render(plain[start:end]))
		i = end
	}
	return b.String()
}
