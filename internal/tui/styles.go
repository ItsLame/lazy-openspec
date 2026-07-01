package tui

import "github.com/charmbracelet/lipgloss"

// Style palette. Colors degrade automatically on NO_COLOR terminals via termenv.
var (
	colBorder  = lipgloss.Color("240")
	colFocus   = lipgloss.Color("39")
	colMuted   = lipgloss.Color("246")
	colTitle   = lipgloss.Color("39")
	colDraft   = lipgloss.Color("246")
	colActive  = lipgloss.Color("214")
	colDone    = lipgloss.Color("42")
	colErr     = lipgloss.Color("203")
	colSelBg   = lipgloss.Color("236")
	colSelFg   = lipgloss.Color("255")
	colHintKey = lipgloss.Color("39")
)

var (
	titleFocused = lipgloss.NewStyle().Bold(true).Foreground(colTitle)
	titleBlur    = lipgloss.NewStyle().Bold(true).Foreground(colMuted)

	selectedItem = lipgloss.NewStyle().Background(colSelBg).Foreground(colSelFg).Bold(true)
	normalItem   = lipgloss.NewStyle()
	mutedText    = lipgloss.NewStyle().Foreground(colMuted)
	errText      = lipgloss.NewStyle().Foreground(colErr).Bold(true)

	glyphDraft  = lipgloss.NewStyle().Foreground(colDraft).Render("○")
	glyphActive = lipgloss.NewStyle().Foreground(colActive).Render("◉")
	glyphDone   = lipgloss.NewStyle().Foreground(colDone).Render("✓")

	tabActive   = lipgloss.NewStyle().Bold(true).Foreground(colTitle).Underline(true)
	tabInactive = lipgloss.NewStyle().Foreground(colMuted)

	hintKey  = lipgloss.NewStyle().Foreground(colHintKey)
	hintDesc = lipgloss.NewStyle().Foreground(colMuted)

	logTitle = lipgloss.NewStyle().Bold(true).Foreground(colMuted)
)

// small inline style helpers.
func lipglossBold(s string) string { return lipgloss.NewStyle().Bold(true).Render(s) }
func lipglossColor(c, s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(c)).Render(s)
}
func faint(s string) string { return lipgloss.NewStyle().Faint(true).Render(s) }

// borderColor picks the panel border color based on focus.
func borderColor(focused bool) lipgloss.Color {
	if focused {
		return colFocus
	}
	return colBorder
}

// panelBox renders a titled, bordered panel sized to (w, h) total cells. Body is
// expected to already fit; overflow is clipped via MaxHeight/MaxWidth.
func panelBox(title, body string, w, h int, focused bool) string {
	if w < 4 {
		w = 4
	}
	if h < 3 {
		h = 3
	}
	ts := titleBlur
	if focused {
		ts = titleFocused
	}
	inner := ts.Render(title) + "\n" + body
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor(focused)).
		Width(w - 2).
		Height(h - 2).
		MaxWidth(w).
		MaxHeight(h).
		Render(inner)
}
