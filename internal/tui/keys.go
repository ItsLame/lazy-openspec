package tui

// binding is a single (keys, description) pair for the hint bar and help overlay.
type binding struct {
	keys string
	desc string
}

// dashboardHelp lists keybindings available on the dashboard.
var dashboardHelp = []binding{
	{"↑/↓ j/k", "move selection"},
	{"tab / 1-3", "switch panel"},
	{"enter", "open"},
	{"r", "refresh"},
	{"v", "validate"},
	{"a", "apply instructions"},
	{"A", "archive"},
	{"x", "actions menu"},
	{"?", "help"},
	{"q", "quit"},
}

// changeDetailHelp lists keybindings in the change detail view.
var changeDetailHelp = []binding{
	{"[ / ]  ←/→", "switch artifact"},
	{"↑/↓ j/k", "scroll"},
	{"space", "toggle task (tasks tab)"},
	{"v", "validate"},
	{"a", "apply instructions"},
	{"A", "archive"},
	{"esc", "back"},
	{"?", "help"},
	{"q", "quit"},
}

// specDetailHelp lists keybindings in the spec detail view.
var specDetailHelp = []binding{
	{"↑/↓ j/k", "scroll"},
	{"n / p", "next / prev requirement"},
	{"esc", "back"},
	{"?", "help"},
	{"q", "quit"},
}

// helpFor returns the help entries for the current screen.
func helpFor(s screen) []binding {
	switch s {
	case screenChangeDetail:
		return changeDetailHelp
	case screenSpecDetail:
		return specDetailHelp
	default:
		return dashboardHelp
	}
}
