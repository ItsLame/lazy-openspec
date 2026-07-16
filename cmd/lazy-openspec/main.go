// Command lazy-openspec is a lazygit-style terminal UI for OpenSpec. It browses
// changes and specs, renders artifacts readably, toggles tasks, and runs the
// openspec workflow commands — all by shelling out to the openspec CLI.
package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/itslame/lazy-openspec/internal/openspec"
	"github.com/itslame/lazy-openspec/internal/tui"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	var (
		store       string
		showVersion bool
	)
	flag.StringVar(&store, "store", "", "target a registered OpenSpec store by id")
	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.Usage = usage
	flag.Parse()

	if showVersion {
		fmt.Printf("lazy-openspec %s\n", version)
		return
	}

	opts := []openspec.Option{}
	if store != "" {
		opts = append(opts, openspec.WithStore(store))
	}
	client := openspec.New(opts...)

	// WithReportFocus lets the TUI refresh when the terminal regains focus, so an
	// agent editing openspec/ from another pane is picked up on switching back.
	// Terminals without DEC mode 1004 (or tmux without `focus-events on`) simply
	// emit no events and fall back to the manual `r` refresh.
	p := tea.NewProgram(tui.New(client), tea.WithAltScreen(), tea.WithReportFocus())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "lazy-openspec:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `lazy-openspec — a lazygit-style TUI for OpenSpec

Usage:
  lazy-openspec [--store <id>]

Flags:
  --store <id>   Target a registered OpenSpec store by id
  --version      Print version and exit
  -h, --help     Show this help

Keys: tab switch panel · ↑↓ move · enter open · [ ] artifact · space toggle task
      v validate · a apply instructions · A archive · x actions · ? help · q quit
`)
}
