package tui

// binding is a single (keys, description) pair for the hint bar and help overlay.
type binding struct {
	keys string
	desc string
}

// navHelp lists keybindings available while the list (nav) pane is focused.
var navHelp = []binding{
	{"↑/↓ j/k", "move selection"},
	{"tab / 1-3", "switch panel"},
	{"enter", "focus preview pane"},
	{"r", "refresh"},
	{"v", "validate"},
	{"a", "apply instructions"},
	{"A", "archive"},
	{"x", "actions menu"},
	{"?", "help"},
	{"q", "quit"},
}

// previewHelp lists keybindings available while the preview pane is focused.
var previewHelp = []binding{
	{"↑/↓ j/k", "scroll"},
	{"ctrl+d / ctrl+u", "half page"},
	{"g / G", "top / bottom"},
	{"[ / ]", "switch tab / requirement"},
	{"space", "toggle task (tasks tab)"},
	{"/", "search"},
	{"n / N", "next / prev match"},
	{"esc", "clear search / back to list"},
	{"?", "help"},
	{"q", "quit"},
}

// helpEntries returns the help entries for the active pane.
func (m Model) helpEntries() []binding {
	if m.activePane == panePreview {
		return previewHelp
	}
	return navHelp
}
