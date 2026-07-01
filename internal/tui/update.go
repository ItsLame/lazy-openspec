package tui

import (
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/itslame/lazy-openspec/internal/openspec"
	"github.com/itslame/lazy-openspec/internal/tasks"
)

// Update handles all incoming messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.ready = true
		d := m.layout()
		m.vp.Width, m.vp.Height = d.vpW, d.vpH
		m.md.SetWidth(d.vpW)
		m = m.refreshMain()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case changesMsg:
		if msg.err != nil {
			m.loadErr = msg.err
			m = m.refreshMain()
			return m, nil
		}
		m.loadErr = nil
		m.changes = sortChanges(msg.list.Changes)
		m.rootPath = msg.list.Root.Path
		m.clampSel()
		var cmds []tea.Cmd
		cmds = append(cmds, loadArchived(m.rootPath))
		if c, ok := m.selectedChange(); ok {
			cmds = append(cmds, loadStatus(m.client, c.Name))
		}
		m = m.refreshMain()
		return m, tea.Batch(cmds...)

	case specsMsg:
		if msg.err == nil {
			m.specs = msg.list.Specs
			m.clampSel()
		}
		m = m.refreshMain()
		return m, nil

	case archivedMsg:
		m.archived = msg.items
		m.clampSel()
		m = m.refreshMain()
		return m, nil

	case statusMsg:
		if msg.err == nil {
			m.statusCache[msg.change] = msg.st
		}
		m = m.refreshMain()
		return m, nil

	case changeDetailMsg:
		if msg.err == nil {
			m.changeDetail[msg.change] = msg.detail
		}
		m = m.refreshMain()
		return m, nil

	case artifactMsg:
		if msg.err == nil {
			m.detailCache[cacheKey(msg.change, msg.tab)] = msg.content
		}
		m = m.refreshMain()
		return m, nil

	case specDetailMsg:
		if msg.err == nil {
			d := msg.detail
			m.specDetail = &d
			m.reqIdx = 0
		}
		m = m.refreshMain()
		return m, nil

	case logLineMsg:
		m.logs = append(m.logs, msg.line)
		m.trimLogs()
		return m, waitForMsg(m.cmdCh)

	case cmdDoneMsg:
		m.running = false
		if msg.err != nil {
			m.logs = append(m.logs, errText.Render("✗ "+msg.label+" failed: "+msg.err.Error()))
		} else {
			m.logs = append(m.logs, glyphDone+" "+msg.label+" completed")
		}
		m.trimLogs()
		// Refresh data after a mutation.
		m.client.Invalidate()
		return m, tea.Batch(loadChanges(m.client), loadSpecs(m.client), loadArchived(m.rootPath))
	}

	// Delegate remaining messages to the viewport (mouse, etc.).
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

// handleKey routes key presses, honoring overlays and the current screen.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()

	// Global quit.
	if k == "ctrl+c" {
		m.quitting = true
		return m, tea.Quit
	}

	// Confirm prompt takes priority.
	if m.confirm != nil {
		switch k {
		case "y", "Y", "enter":
			cmd := m.confirm.onYes(&m)
			m.confirm = nil
			return m, cmd
		case "n", "N", "esc", "q":
			m.confirm = nil
			return m, nil
		}
		return m, nil
	}

	// Help overlay.
	if m.showHelp {
		if k == "?" || k == "esc" || k == "q" {
			m.showHelp = false
		}
		return m, nil
	}

	// Actions overlay.
	if m.showActions {
		switch k {
		case "esc", "x", "q":
			m.showActions = false
			return m, nil
		case "v", "a", "A":
			m.showActions = false
			return m.runAction(k)
		}
		return m, nil
	}

	// Common bindings across screens.
	switch k {
	case "q":
		m.quitting = true
		return m, tea.Quit
	case "?":
		m.showHelp = true
		return m, nil
	case "r":
		m.client.Invalidate()
		return m, tea.Batch(loadChanges(m.client), loadSpecs(m.client), loadArchived(m.rootPath))
	case "v", "a", "A":
		return m.runAction(k)
	case "x":
		m.showActions = true
		return m, nil
	}

	switch m.screen {
	case screenDashboard:
		return m.handleDashboardKey(k)
	case screenChangeDetail:
		return m.handleChangeDetailKey(k)
	case screenSpecDetail:
		return m.handleSpecDetailKey(k)
	}
	return m, nil
}

func (m Model) handleDashboardKey(k string) (tea.Model, tea.Cmd) {
	switch k {
	case "tab", "l":
		m.focus = (m.focus + 1) % numPanels
	case "shift+tab", "h":
		m.focus = (m.focus + numPanels - 1) % numPanels
	case "1":
		m.focus = panelChanges
	case "2":
		m.focus = panelSpecs
	case "3":
		m.focus = panelArchive
	case "up", "k":
		m.moveSel(-1)
	case "down", "j":
		m.moveSel(1)
	case "enter":
		return m.enterSelected()
	default:
		return m, nil
	}
	// Selection or focus changed: refresh preview and load status if needed.
	var cmd tea.Cmd
	if c, ok := m.selectedChange(); ok {
		if _, cached := m.statusCache[c.Name]; !cached {
			cmd = loadStatus(m.client, c.Name)
		}
	}
	m = m.refreshMain()
	return m, cmd
}

func (m Model) handleChangeDetailKey(k string) (tea.Model, tea.Cmd) {
	switch k {
	case "esc":
		m.screen = screenDashboard
		m = m.refreshMain()
		return m, nil
	case "]", "right", "l":
		m.tab = (m.tab + 1) % numTabs
		return m.afterTabChange()
	case "[", "left", "h":
		m.tab = (m.tab + numTabs - 1) % numTabs
		return m.afterTabChange()
	case " ":
		return m.toggleTask()
	case "up", "k":
		if m.tab == tabTasks {
			m.moveTaskCursor(-1)
			m = m.refreshMain()
			return m, nil
		}
	case "down", "j":
		if m.tab == tabTasks {
			m.moveTaskCursor(1)
			m = m.refreshMain()
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(keyMsg(k))
	return m, cmd
}

// moveTaskCursor moves the task selection within the tasks tab.
func (m *Model) moveTaskCursor(delta int) {
	flat := flatTasks(tasks.Parse(m.detailCache[cacheKey(m.curChange, tabTasks)]))
	n := len(flat)
	if n == 0 {
		return
	}
	m.taskCursor += delta
	if m.taskCursor < 0 {
		m.taskCursor = 0
	}
	if m.taskCursor >= n {
		m.taskCursor = n - 1
	}
}

func (m Model) handleSpecDetailKey(k string) (tea.Model, tea.Cmd) {
	switch k {
	case "esc":
		m.screen = screenDashboard
		m = m.refreshMain()
		return m, nil
	case "n":
		if len(m.reqOffsets) > 0 {
			m.reqIdx = (m.reqIdx + 1) % len(m.reqOffsets)
			m.vp.SetYOffset(m.reqOffsets[m.reqIdx])
		}
		return m, nil
	case "p":
		if len(m.reqOffsets) > 0 {
			m.reqIdx = (m.reqIdx + len(m.reqOffsets) - 1) % len(m.reqOffsets)
			m.vp.SetYOffset(m.reqOffsets[m.reqIdx])
		}
		return m, nil
	default:
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(keyMsg(k))
		return m, cmd
	}
}

// enterSelected opens the highlighted change or spec.
func (m Model) enterSelected() (tea.Model, tea.Cmd) {
	if c, ok := m.selectedChange(); ok {
		archived := m.focus == panelArchive
		return m.openChange(c.Name, archived)
	}
	if s, ok := m.selectedSpec(); ok {
		m.screen = screenSpecDetail
		m.curSpec = s.Name
		m.specDetail = nil
		m.reqIdx = 0
		m.vp.GotoTop()
		m = m.refreshMain()
		return m, loadSpecDetail(m.client, s.Name)
	}
	return m, nil
}

// openChange enters the change detail view and loads its artifacts.
func (m Model) openChange(name string, archived bool) (tea.Model, tea.Cmd) {
	m.screen = screenChangeDetail
	m.curChange = name
	m.curArchived = archived
	m.tab = tabProposal
	m.taskCursor = 0
	sub := "changes"
	if archived {
		sub = "changes/archive"
	}
	m.curChangeDir = openspec.ArtifactPath(m.rootPath, "openspec/"+sub+"/"+name)
	m.vp.GotoTop()
	m = m.refreshMain()
	return m, tea.Batch(
		loadChangeDetail(m.client, name),
		loadArtifact(m.curChangeDir, name, tabProposal),
	)
}

// afterTabChange loads the newly selected tab's content if not cached.
func (m Model) afterTabChange() (tea.Model, tea.Cmd) {
	m.vp.GotoTop()
	m.taskCursor = 0
	var cmd tea.Cmd
	if m.tab == tabProposal || m.tab == tabDesign || m.tab == tabTasks {
		if _, ok := m.detailCache[cacheKey(m.curChange, m.tab)]; !ok {
			cmd = loadArtifact(m.curChangeDir, m.curChange, m.tab)
		}
	} else if m.tab == tabSpecs {
		if _, ok := m.changeDetail[m.curChange]; !ok {
			cmd = loadChangeDetail(m.client, m.curChange)
		}
	}
	m = m.refreshMain()
	return m, cmd
}

// toggleTask flips the selected task in the tasks tab and persists it.
func (m Model) toggleTask() (tea.Model, tea.Cmd) {
	if m.tab != tabTasks || m.curArchived {
		return m, nil
	}
	key := cacheKey(m.curChange, tabTasks)
	content := m.detailCache[key]
	flat := flatTasks(tasks.Parse(content))
	if len(flat) == 0 {
		m.logs = append(m.logs, mutedText.Render("no tasks to toggle"))
		return m, nil
	}
	if m.taskCursor >= len(flat) {
		m.taskCursor = len(flat) - 1
	}
	num := flat[m.taskCursor].Number
	newContent, _, err := tasks.Toggle(content, num)
	if err != nil {
		m.logs = append(m.logs, errText.Render("✗ "+err.Error()))
		m.trimLogs()
		return m, nil
	}
	if err := writeFile(m.curChangeDir+"/tasks.md", newContent); err != nil {
		m.logs = append(m.logs, errText.Render("✗ write tasks.md: "+err.Error()))
		m.trimLogs()
		return m, nil
	}
	m.detailCache[key] = newContent
	m.client.Invalidate()
	m = m.refreshMain()
	return m, tea.Batch(loadStatus(m.client, m.curChange), loadChanges(m.client))
}

// runAction dispatches v/a/A to the corresponding openspec command.
func (m Model) runAction(k string) (tea.Model, tea.Cmd) {
	name, ok := m.actionTarget()
	if !ok {
		m.logs = append(m.logs, mutedText.Render("select a change first"))
		return m, nil
	}
	store := m.client.Store()
	switch k {
	case "v":
		args := withStore([]string{"validate", name, "--no-interactive"}, store)
		return m.startCommand("validate "+name, args)
	case "a":
		// There is no `openspec apply` CLI command (apply is the AI skill); show
		// the apply instructions, which is the real, useful affordance here.
		args := withStore([]string{"instructions", "apply", "--change", name}, store)
		return m.startCommand("instructions apply "+name, args)
	case "A":
		m.confirm = &confirmState{
			prompt: "Archive '" + name + "'? This updates main specs. (y/n)",
			onYes: func(mm *Model) tea.Cmd {
				args := withStore([]string{"archive", name, "-y"}, store)
				mdl, cmd := mm.startCommand("archive "+name, args)
				*mm = mdl.(Model)
				return cmd
			},
		}
		return m, nil
	}
	return m, nil
}

// startCommand begins a streaming openspec invocation.
func (m Model) startCommand(label string, args []string) (tea.Model, tea.Cmd) {
	if m.running {
		m.logs = append(m.logs, mutedText.Render("a command is already running"))
		return m, nil
	}
	m.running = true
	m.logs = append(m.logs, hintKey.Render("$ openspec "+strings.Join(args, " ")))
	m.trimLogs()
	return m, runProcess("openspec", args, label, m.cmdCh)
}

// actionTarget returns the change an action should apply to (detail view's
// change, or the dashboard selection).
func (m Model) actionTarget() (string, bool) {
	if m.screen == screenChangeDetail && m.curChange != "" {
		return m.curChange, true
	}
	if c, ok := m.selectedChange(); ok {
		return c.Name, true
	}
	return "", false
}

// ---- selection helpers ------------------------------------------------------

func (m *Model) moveSel(delta int) {
	n := m.panelLen(m.focus)
	if n == 0 {
		return
	}
	m.sel[m.focus] += delta
	if m.sel[m.focus] < 0 {
		m.sel[m.focus] = 0
	}
	if m.sel[m.focus] >= n {
		m.sel[m.focus] = n - 1
	}
}

func (m Model) panelLen(p panel) int {
	switch p {
	case panelChanges:
		return len(m.changes)
	case panelSpecs:
		return len(m.specs)
	case panelArchive:
		return len(m.archived)
	}
	return 0
}

func (m *Model) clampSel() {
	for p := panel(0); p < numPanels; p++ {
		n := m.panelLen(p)
		if m.sel[p] >= n {
			m.sel[p] = maxZero(n - 1)
		}
		if m.sel[p] < 0 {
			m.sel[p] = 0
		}
	}
}

func (m *Model) trimLogs() {
	const maxLogs = 200
	if len(m.logs) > maxLogs {
		m.logs = m.logs[len(m.logs)-maxLogs:]
	}
}

func cacheKey(change string, tab artifactTab) string {
	return change + "/" + tabNames[tab]
}

func withStore(args []string, store string) []string {
	if store == "" {
		return args
	}
	return append(args, "--store", store)
}

// sortChanges orders changes by lifecycle (active, draft, completed) then name,
// so the grouped panel and the selection index stay aligned.
func sortChanges(in []openspec.ChangeSummary) []openspec.ChangeSummary {
	out := append([]openspec.ChangeSummary(nil), in...)
	rank := func(c openspec.ChangeSummary) int {
		switch c.Lifecycle() {
		case openspec.Active:
			return 0
		case openspec.Draft:
			return 1
		default:
			return 2
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if ri, rj := rank(out[i]), rank(out[j]); ri != rj {
			return ri < rj
		}
		return out[i].Name < out[j].Name
	})
	return out
}

// flatTasks flattens groups into a single task slice in document order, matching
// the order the tasks tab renders them (used to map the task cursor).
func flatTasks(groups []tasks.Group) []tasks.Task {
	var out []tasks.Task
	for _, g := range groups {
		out = append(out, g.Tasks...)
	}
	return out
}

func maxZero(n int) int {
	if n < 0 {
		return 0
	}
	return n
}

// keyMsg builds a synthetic KeyMsg so scroll keys can be forwarded to the
// viewport's own Update.
func keyMsg(k string) tea.KeyMsg {
	switch k {
	case "up", "k":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down", "j":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "pgup":
		return tea.KeyMsg{Type: tea.KeyPgUp}
	case "pgdown", "pgdn":
		return tea.KeyMsg{Type: tea.KeyPgDown}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}
