// Package tasks parses and edits an OpenSpec change's tasks.md. It provides a
// structured view (groups + tasks) for rendering and a surgical Toggle that
// flips a single checkbox marker while preserving every other byte of the file.
package tasks

import (
	"fmt"
	"regexp"
	"strings"
)

// taskLine matches "- [ ] 1.2 description" (or [x]/[X]), capturing indent,
// state, number and text.
var taskLine = regexp.MustCompile(`^(\s*)-\s*\[([ xX])\]\s+(\d+(?:\.\d+)*)\s+(.*)$`)

// groupLine matches "## 1. Group title".
var groupLine = regexp.MustCompile(`^##\s+(\d+)\.\s*(.*)$`)

// Task is a single checkbox item.
type Task struct {
	Number string // e.g. "1.3"
	Text   string // description text after the number
	Done   bool
	Line   int // 0-based line index within the file
}

// Group is a "## N. Title" heading and the tasks beneath it.
type Group struct {
	Number string
	Title  string
	Tasks  []Task
}

// Completed returns the number of done tasks in the group.
func (g Group) Completed() int {
	n := 0
	for _, t := range g.Tasks {
		if t.Done {
			n++
		}
	}
	return n
}

// Total returns the number of tasks in the group.
func (g Group) Total() int { return len(g.Tasks) }

// Parse turns tasks.md content into groups. Tasks that appear before any group
// heading are collected under an untitled group.
func Parse(content string) []Group {
	var groups []Group
	var cur *Group
	ensure := func() *Group {
		if cur == nil {
			groups = append(groups, Group{})
			cur = &groups[len(groups)-1]
		}
		return cur
	}
	for i, line := range strings.Split(content, "\n") {
		if m := groupLine.FindStringSubmatch(line); m != nil {
			groups = append(groups, Group{Number: m[1], Title: strings.TrimSpace(m[2])})
			cur = &groups[len(groups)-1]
			continue
		}
		if m := taskLine.FindStringSubmatch(line); m != nil {
			g := ensure()
			g.Tasks = append(g.Tasks, Task{
				Number: m[3],
				Text:   strings.TrimSpace(m[4]),
				Done:   m[2] == "x" || m[2] == "X",
				Line:   i,
			})
		}
	}
	return groups
}

// Progress returns (completed, total) across all tasks.
func Progress(content string) (completed, total int) {
	for _, g := range Parse(content) {
		completed += g.Completed()
		total += g.Total()
	}
	return
}

// Toggle flips the checkbox for the task whose number matches. It changes only
// that line's marker (`[ ]` <-> `[x]`) and preserves all other bytes. It returns
// the new content and the task's new done-state. If no matching task is found
// (e.g. the file changed on disk), it returns an error so the caller can prompt
// for a refresh.
func Toggle(content, number string) (string, bool, error) {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		m := taskLine.FindStringSubmatch(line)
		if m == nil || m[3] != number {
			continue
		}
		done := m[2] == "x" || m[2] == "X"
		if done {
			lines[i] = strings.Replace(line, "["+m[2]+"]", "[ ]", 1)
		} else {
			lines[i] = strings.Replace(line, "[ ]", "[x]", 1)
		}
		return strings.Join(lines, "\n"), !done, nil
	}
	return content, false, fmt.Errorf("task %s not found (file may have changed on disk)", number)
}
