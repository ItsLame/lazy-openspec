package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// Style palette. Colors are drawn from the terminal's ANSI 16-color palette
// (indices 0-15) and the terminal default (unset) foreground, so the interface
// follows the user's terminal theme like lazygit/lazydocker rather than pinning
// fixed xterm-256 colors. Colors degrade automatically on NO_COLOR terminals via
// termenv. colNone is the unset/"default" color: it inherits the terminal's own
// foreground.
var (
	colNone   = lipgloss.NoColor{}   // terminal default (inherits fg/bg)
	colFocus  = lipgloss.Color("2")  // active border: green
	colTitle  = lipgloss.Color("6")  // titles/accents: cyan
	colActive = lipgloss.Color("3")  // active/in-progress: yellow
	colDone   = lipgloss.Color("2")  // done: green
	colErr    = lipgloss.Color("1")  // error: red
	colSelBg  = lipgloss.Color("4")  // selected line background: blue
)

var (
	titleFocused = lipgloss.NewStyle().Bold(true).Foreground(colTitle)
	titleBlur    = lipgloss.NewStyle().Bold(true).Faint(true)

	// Selected line: blue background with the terminal default foreground, matching
	// lazygit's selectedLineBgColor.
	selectedItem = lipgloss.NewStyle().Background(colSelBg).Foreground(colNone).Bold(true)
	normalItem   = lipgloss.NewStyle()
	// Muted text inherits the terminal foreground rendered faint, so it stays
	// legible on any theme instead of forcing a fixed grey.
	mutedText = lipgloss.NewStyle().Faint(true)
	errText   = lipgloss.NewStyle().Foreground(colErr).Bold(true)

	glyphDraft  = lipgloss.NewStyle().Faint(true).Render("○")
	glyphActive = lipgloss.NewStyle().Foreground(colActive).Render("◉")
	glyphDone   = lipgloss.NewStyle().Foreground(colDone).Render("✓")

	tabActive   = lipgloss.NewStyle().Bold(true).Foreground(colTitle).Underline(true)
	tabInactive = lipgloss.NewStyle().Faint(true)

	hintKey  = lipgloss.NewStyle().Foreground(colTitle)
	hintDesc = lipgloss.NewStyle().Faint(true)

	logTitle = lipgloss.NewStyle().Bold(true).Faint(true)
)

// small inline style helpers.
func lipglossBold(s string) string { return lipgloss.NewStyle().Bold(true).Render(s) }
func lipglossColor(c, s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(c)).Render(s)
}
func faint(s string) string { return lipgloss.NewStyle().Faint(true).Render(s) }

// borderColor picks the panel border color based on focus. An unfocused panel
// uses the terminal default so inactive borders match the user's theme.
func borderColor(focused bool) lipgloss.TerminalColor {
	if focused {
		return colFocus
	}
	return colNone
}

// panelBox renders a bordered panel sized to (w, h) total cells with its title
// embedded in the top border, lazygit-style. Body is expected to already fit;
// overflow is clipped via MaxHeight/MaxWidth.
func panelBox(title, body string, w, h int, focused bool) string {
	if w < 4 {
		w = 4
	}
	if h < 3 {
		h = 3
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor(focused)).
		Padding(0, 1).
		Width(w - 2).
		Height(h - 2).
		MaxWidth(w).
		MaxHeight(h).
		Render(body)
	return withBorderTitle(box, borderTitleStyle(focused).Render(title), focused)
}

// borderTitleStyle styles a title embedded in a box frame: the active border
// colour when the box is focused, so the whole top line reads as one piece the
// way lazygit draws it, and de-emphasised otherwise.
func borderTitleStyle(focused bool) lipgloss.Style {
	if focused {
		return lipgloss.NewStyle().Bold(true).Foreground(colFocus)
	}
	return titleBlur
}

// withBorderTitle rewrites a rendered box's top border so the (already styled)
// title passes through the frame instead of occupying a body row, lazygit-style:
// ╭─[1]─Changes───╮. A title too wide for the frame is stripped of its styling and
// truncated so it can never break the border.
func withBorderTitle(box, title string, focused bool) string {
	parts := strings.SplitN(box, "\n", 2)
	avail := lipgloss.Width(parts[0]) - 3 // corners plus the lead-in dash
	if avail < 1 {
		return box
	}
	if lipgloss.Width(title) > avail {
		title = trunc(ansi.Strip(title), avail)
	}
	frame := lipgloss.NewStyle().Foreground(borderColor(focused))
	top := frame.Render("╭─") + title + frame.Render(strings.Repeat("─", avail-lipgloss.Width(title))+"╮")
	if len(parts) == 1 {
		return top
	}
	return top + "\n" + parts[1]
}
