// Package render turns OpenSpec artifacts into readable terminal output. Prose
// goes through Glamour; OpenSpec's structured blocks (requirements, scenarios,
// task checklists) are rendered semantically so they read better than raw
// markdown.
package render

import (
	"os"
	"strings"
	"sync"

	"github.com/charmbracelet/glamour"
)

// minWidth guards Glamour against absurdly narrow wrap widths.
const minWidth = 20

// Markdown wraps a Glamour renderer, rebuilding it when the target width or
// color mode changes.
type Markdown struct {
	mu    sync.Mutex
	width int
	mono  bool
	r     *glamour.TermRenderer
}

// NewMarkdown builds a renderer for the given width. It selects a monochrome
// style when the terminal has no color support (NO_COLOR set), satisfying the
// no-color readability requirement.
func NewMarkdown(width int) *Markdown {
	m := &Markdown{mono: noColor()}
	m.SetWidth(width)
	return m
}

// noColor reports whether color output should be suppressed.
func noColor() bool {
	return os.Getenv("NO_COLOR") != ""
}

// SetWidth rebuilds the underlying renderer for a new wrap width. It is a no-op
// when the width is unchanged.
func (m *Markdown) SetWidth(width int) {
	if width < minWidth {
		width = minWidth
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.r != nil && width == m.width {
		return
	}
	m.width = width
	opts := []glamour.TermRendererOption{glamour.WithWordWrap(width)}
	if m.mono {
		opts = append(opts, glamour.WithStandardStyle("notty"))
	} else {
		opts = append(opts, glamour.WithAutoStyle())
	}
	// glamour only errors on invalid style names, which we never pass; on the
	// off chance it does, fall back to leaving the previous renderer in place.
	if r, err := glamour.NewTermRenderer(opts...); err == nil {
		m.r = r
	}
}

// Render renders markdown to styled, word-wrapped terminal text. On any failure
// it falls back to the raw markdown so content is never lost.
func (m *Markdown) Render(md string) string {
	m.mu.Lock()
	r := m.r
	m.mu.Unlock()
	if r == nil {
		return md
	}
	out, err := r.Render(md)
	if err != nil {
		return md
	}
	return strings.TrimRight(out, "\n")
}
