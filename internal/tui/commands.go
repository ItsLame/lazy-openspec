package tui

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/itslame/lazy-openspec/internal/openspec"
	"github.com/itslame/lazy-openspec/internal/tasks"
)

// ---- messages ---------------------------------------------------------------

type changesMsg struct {
	list openspec.ChangeList
	err  error
}
type specsMsg struct {
	list openspec.SpecList
	err  error
}
type statusMsg struct {
	change string
	st     openspec.Status
	err    error
}
type changeDetailMsg struct {
	change string
	detail openspec.ChangeDetail
	err    error
}
type artifactMsg struct {
	change  string
	tab     artifactTab
	content string
	err     error
}
type specDetailMsg struct {
	id     string
	detail openspec.SpecDetail
	err    error
}
type archivedMsg struct {
	items []openspec.ChangeSummary
}

// archivedOverview is the on-disk overview of an archived change, used in place
// of `openspec status`/`show` (which cannot resolve archived names).
type archivedOverview struct {
	completed, total                 int
	hasProposal, hasDesign, hasTasks bool
}
type archivedOverviewMsg struct {
	change string
	ov     archivedOverview
}
type logLineMsg struct{ line string }
type cmdDoneMsg struct {
	label string
	err   error
}

// ---- data commands ----------------------------------------------------------

func loadChanges(c *openspec.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), openspec.DefaultTimeout)
		defer cancel()
		list, err := c.Changes(ctx)
		return changesMsg{list: list, err: err}
	}
}

func loadSpecs(c *openspec.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), openspec.DefaultTimeout)
		defer cancel()
		list, err := c.Specs(ctx)
		return specsMsg{list: list, err: err}
	}
}

func loadStatus(c *openspec.Client, change string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), openspec.DefaultTimeout)
		defer cancel()
		st, err := c.Status(ctx, change)
		return statusMsg{change: change, st: st, err: err}
	}
}

func loadChangeDetail(c *openspec.Client, change string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), openspec.DefaultTimeout)
		defer cancel()
		d, err := c.ChangeDetail(ctx, change)
		return changeDetailMsg{change: change, detail: d, err: err}
	}
}

func loadSpecDetail(c *openspec.Client, id string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), openspec.DefaultTimeout)
		defer cancel()
		d, err := c.SpecDetail(ctx, id)
		return specDetailMsg{id: id, detail: d, err: err}
	}
}

// loadArtifact reads an artifact markdown file (proposal/design/tasks) from a
// change directory. The specs tab is sourced from ChangeDetail, not a file.
func loadArtifact(changeDir, change string, tab artifactTab) tea.Cmd {
	return func() tea.Msg {
		var rel string
		switch tab {
		case tabProposal:
			rel = "proposal.md"
		case tabDesign:
			rel = "design.md"
		case tabTasks:
			rel = "tasks.md"
		default:
			return artifactMsg{change: change, tab: tab}
		}
		content, err := readFile(filepath.Join(changeDir, rel))
		return artifactMsg{change: change, tab: tab, content: content, err: err}
	}
}

// loadArchivedOverview derives an archived change's overview from disk: task
// counts parsed from tasks.md and which artifact files are present. Archived
// changes are not resolvable by the CLI, so this never shells out.
func loadArchivedOverview(changeDir, change string) tea.Cmd {
	return func() tea.Msg {
		tasksContent, _ := readFile(filepath.Join(changeDir, "tasks.md"))
		ov := archivedOverview{}
		for _, g := range tasks.Parse(tasksContent) {
			ov.completed += g.Completed()
			ov.total += g.Total()
		}
		exists := func(name string) bool {
			_, err := os.Stat(filepath.Join(changeDir, name))
			return err == nil
		}
		ov.hasProposal = exists("proposal.md")
		ov.hasDesign = exists("design.md")
		ov.hasTasks = exists("tasks.md")
		return archivedOverviewMsg{change: change, ov: ov}
	}
}

// loadArchived lists archived change directories under the resolved root.
func loadArchived(rootPath string) tea.Cmd {
	return func() tea.Msg {
		dir := filepath.Join(rootPath, "openspec", "changes", "archive")
		names := listDirs(dir)
		items := make([]openspec.ChangeSummary, 0, len(names))
		for _, n := range names {
			items = append(items, openspec.ChangeSummary{Name: n})
		}
		return archivedMsg{items: items}
	}
}

// ---- streaming command runner ----------------------------------------------

// runProcess starts a process and streams its combined output line-by-line onto
// ch, followed by a cmdDoneMsg. The returned tea.Cmd delivers the first buffered
// message; the Update loop keeps calling waitForMsg to drain the rest.
func runProcess(name string, args []string, label string, ch chan tea.Msg) tea.Cmd {
	go func() {
		c := exec.Command(name, args...)
		pr, pw := io.Pipe()
		c.Stdout = pw
		c.Stderr = pw
		if err := c.Start(); err != nil {
			ch <- cmdDoneMsg{label: label, err: err}
			return
		}
		done := make(chan error, 1)
		go func() {
			done <- c.Wait()
			_ = pw.Close()
		}()
		sc := bufio.NewScanner(pr)
		sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for sc.Scan() {
			ch <- logLineMsg{line: sc.Text()}
		}
		ch <- cmdDoneMsg{label: label, err: <-done}
	}()
	return waitForMsg(ch)
}

// waitForMsg blocks on the channel and returns the next streamed message.
func waitForMsg(ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg { return <-ch }
}
