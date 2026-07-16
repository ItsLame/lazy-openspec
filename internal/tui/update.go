package tui

import (
	"sort"
	"strings"
	"time"

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

	case tea.BlurMsg:
		// Stale-while-blurred: note the blur and do nothing else — no polling, no
		// subprocesses while nobody is looking.
		m.blurred = true
		return m, nil

	case tea.FocusMsg:
		// Refresh on regaining focus, but only when returning from a real blur
		// (some terminals emit a focus event just from enabling reporting), never
		// while a streaming command is running (its completion refreshes anyway),
		// and at most once per debounce window.
		if !m.blurred {
			return m, nil
		}
		m.blurred = false
		if m.running || time.Since(m.lastRefresh) < refreshDebounce {
			return m, nil
		}
		cmd := m.refreshAll()
		m = m.refreshMain()
		return m, cmd

	case changesMsg:
		if msg.err != nil {
			m.loadErr = msg.err
			m = m.refreshMain()
			return m, nil
		}
		m.loadErr = nil
		prev := m.visibleNameAt(panelChanges, m.sel[panelChanges])
		m.changes = sortChanges(msg.list.Changes)
		m.rootPath = msg.list.Root.Path
		m.preserveSel(panelChanges, prev)
		m.syncSelection()
		cmds := []tea.Cmd{loadArchived(m.rootPath)}
		if c := m.ensurePreviewLoaded(); c != nil {
			cmds = append(cmds, c)
		}
		m = m.refreshMain()
		return m, tea.Batch(cmds...)

	case specsMsg:
		if msg.err == nil {
			prev := m.visibleNameAt(panelSpecs, m.sel[panelSpecs])
			m.specs = msg.list.Specs
			m.preserveSel(panelSpecs, prev)
		}
		m.syncSelection()
		cmd := m.ensurePreviewLoaded()
		m = m.refreshMain()
		return m, cmd

	case archivedMsg:
		prev := m.visibleNameAt(panelArchive, m.sel[panelArchive])
		m.archived = msg.items
		m.preserveSel(panelArchive, prev)
		m.syncSelection()
		cmd := m.ensurePreviewLoaded()
		m = m.refreshMain()
		return m, cmd

	case statusMsg:
		if msg.err != nil {
			m.statusErr[msg.change] = msg.err
		} else {
			m.statusCache[msg.change] = msg.st
			delete(m.statusErr, msg.change)
		}
		m = m.refreshMain()
		return m, nil

	case changeDetailMsg:
		if msg.err != nil {
			m.specsErr[msg.change] = msg.err
		} else {
			m.changeDetail[msg.change] = msg.detail
			delete(m.specsErr, msg.change)
		}
		m = m.refreshMain()
		return m, nil

	case artifactMsg:
		key := cacheKey(msg.change, msg.tab)
		if msg.err != nil {
			m.detailErr[key] = msg.err.Error()
		} else {
			m.detailCache[key] = msg.content
			delete(m.detailErr, key)
		}
		m = m.refreshMain()
		return m, nil

	case archivedOverviewMsg:
		m.archivedOv[msg.change] = msg.ov
		m = m.refreshMain()
		return m, nil

	case specDetailMsg:
		if msg.err != nil {
			m.specErr[msg.id] = msg.err
		} else if msg.id == m.curSpec {
			d := msg.detail
			m.specDetail = &d
			m.reqIdx = 0
			delete(m.specErr, msg.id)
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
		// The command mutated the tree: refresh everything, preview included.
		cmd := m.refreshAll()
		m = m.refreshMain()
		return m, cmd
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

	// While typing a search query in the preview, every key feeds the query so
	// that letters like q/r/v are not swallowed by the global bindings below.
	if m.activePane == panePreview && m.search.typing {
		return m.handleSearchInput(msg)
	}

	// Likewise for a list filter query: without this guard, typing a query
	// containing q would quit and one containing A would archive.
	if m.activePane == paneNav && m.filter.typing {
		return m.handleFilterInput(msg)
	}

	// Common bindings across panes.
	switch k {
	case "q":
		m.quitting = true
		return m, tea.Quit
	case "?":
		m.showHelp = true
		return m, nil
	case "r":
		cmd := m.refreshAll()
		m = m.refreshMain()
		return m, cmd
	case "v", "a", "A":
		return m.runAction(k)
	case "x":
		m.showActions = true
		return m, nil
	}

	switch m.activePane {
	case paneNav:
		return m.handleNavKey(k)
	case panePreview:
		return m.handlePreviewKey(msg)
	}
	return m, nil
}

// handleNavKey routes keys while the list pane is focused. Note that left/right
// are deliberately *not* bound to the preview's sections here: the list already
// owns a horizontal axis (h/l cycle panels), so `l` and `→` must not diverge.
func (m Model) handleNavKey(k string) (tea.Model, tea.Cmd) {
	prevFocus := m.focus
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
	case "]":
		return m.previewSection(1)
	case "[":
		return m.previewSection(-1)
	case "/":
		// Filter the focused panel (the preview's own `/` searches its text).
		m.filter = listFilter{panel: m.focus, typing: true}
		m = m.refreshMain()
		return m, nil
	case "esc":
		if !m.filter.active() {
			return m, nil
		}
		prev := m.selectedName()
		m.clearFilter()
		return m.reselectAfterFilter(prev)
	case "enter":
		// Transfer focus to the preview pane; the previewed item is unchanged.
		m.activePane = panePreview
		m = m.refreshMain()
		return m, nil
	default:
		return m, nil
	}
	// Focus left the filtered panel: no panel may stay silently narrowed.
	if m.focus != prevFocus {
		m.clearFilterKeepingSelection()
	}
	// Selection or focus changed: follow the selection and lazily load its preview.
	m.syncSelection()
	cmd := m.ensurePreviewLoaded()
	m = m.refreshMain()
	return m, cmd
}

// previewSection moves the preview to the previous/next section of the selected
// item — the artifact tab of a change, or the requirement of a spec — while
// keyboard focus stays on the list.
func (m Model) previewSection(delta int) (tea.Model, tea.Cmd) {
	if _, ok := m.selectedChange(); ok {
		m.cycleTab(delta)
		return m.afterTabChange()
	}
	if _, ok := m.selectedSpec(); ok {
		m.jumpRequirement(delta)
		return m, nil
	}
	return m, nil
}

// handleFilterInput edits the list-filter query while typing is active. Every
// edit re-selects, so the highlighted row follows the item as the list narrows.
func (m Model) handleFilterInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	prev := m.selectedName()
	switch msg.Type {
	case tea.KeyEnter:
		m.filter.typing = false // confirm: keep the query and the narrowed rows
		m = m.refreshMain()
		return m, nil
	case tea.KeyEsc:
		m.clearFilter()
	case tea.KeyBackspace:
		if r := []rune(m.filter.query); len(r) > 0 {
			m.filter.query = string(r[:len(r)-1])
		}
	case tea.KeySpace:
		m.filter.query += " "
	case tea.KeyRunes:
		m.filter.query += string(msg.Runes)
	default:
		// Not text: hand it to the list, so tab still switches panel (clearing the
		// filter) and the arrows still step between matching rows while typing.
		return m.handleNavKey(msg.String())
	}
	return m.reselectAfterFilter(prev)
}

// reselectAfterFilter restores the selection to prevName's row in the newly
// visible set (falling back to the first row when it no longer matches), then
// lets the preview follow it.
func (m Model) reselectAfterFilter(prevName string) (tea.Model, tea.Cmd) {
	m.sel[m.focus] = m.indexOfVisible(m.focus, prevName)
	m.syncSelection()
	cmd := m.ensurePreviewLoaded()
	m = m.refreshMain()
	return m, cmd
}

// ensurePreviewLoaded dispatches only the loader needed for the current
// selection and active tab, if its data is not already cached (or known to have
// failed). Archived changes are sourced from disk and never hit the CLI.
func (m Model) ensurePreviewLoaded() tea.Cmd {
	if c, ok := m.selectedChange(); ok {
		archived := m.focus == panelArchive
		sub := "changes"
		if archived {
			sub = "changes/archive"
		}
		dir := openspec.ArtifactPath(m.rootPath, "openspec/"+sub+"/"+c.Name)
		var cmds []tea.Cmd
		if archived {
			if _, ok := m.archivedOv[c.Name]; !ok {
				cmds = append(cmds, loadArchivedOverview(dir, c.Name))
			}
			if fileBackedTab(m.tab) {
				key := cacheKey(c.Name, m.tab)
				if _, ok := m.detailCache[key]; !ok {
					if _, failed := m.detailErr[key]; !failed {
						cmds = append(cmds, loadArtifact(dir, c.Name, m.tab))
					}
				}
			}
			return tea.Batch(cmds...)
		}
		switch {
		case m.tab == tabOverview:
			if _, ok := m.statusCache[c.Name]; !ok {
				if _, failed := m.statusErr[c.Name]; !failed {
					cmds = append(cmds, loadStatus(m.client, c.Name))
				}
			}
		case fileBackedTab(m.tab):
			key := cacheKey(c.Name, m.tab)
			if _, ok := m.detailCache[key]; !ok {
				if _, failed := m.detailErr[key]; !failed {
					cmds = append(cmds, loadArtifact(dir, c.Name, m.tab))
				}
			}
		case m.tab == tabSpecs:
			if _, ok := m.changeDetail[c.Name]; !ok {
				if _, failed := m.specsErr[c.Name]; !failed {
					cmds = append(cmds, loadChangeDetail(m.client, c.Name))
				}
			}
		}
		return tea.Batch(cmds...)
	}
	if s, ok := m.selectedSpec(); ok {
		if m.specDetail == nil || m.specDetail.ID != s.Name {
			if _, failed := m.specErr[s.Name]; !failed {
				return loadSpecDetail(m.client, s.Name)
			}
		}
	}
	return nil
}

// ---- refresh ----------------------------------------------------------------

// refreshDebounce is the shortest gap between two focus-driven refreshes, so
// focus flapping (dragging panes around) cannot stack subprocess spawns.
const refreshDebounce = time.Second

// reloadPreview dispatches the loaders for the current selection and active tab
// *unconditionally* — the force-load counterpart of ensurePreviewLoaded, which
// only fills cache misses. This is what makes a refresh re-read what is already
// on screen instead of leaving it stale.
func (m Model) reloadPreview() tea.Cmd {
	if c, ok := m.selectedChange(); ok {
		archived := m.focus == panelArchive
		sub := "changes"
		if archived {
			sub = "changes/archive"
		}
		dir := openspec.ArtifactPath(m.rootPath, "openspec/"+sub+"/"+c.Name)
		var cmds []tea.Cmd
		if archived {
			cmds = append(cmds, loadArchivedOverview(dir, c.Name))
			if fileBackedTab(m.tab) {
				cmds = append(cmds, loadArtifact(dir, c.Name, m.tab))
			}
			return tea.Batch(cmds...)
		}
		switch {
		case m.tab == tabOverview:
			cmds = append(cmds, loadStatus(m.client, c.Name))
		case fileBackedTab(m.tab):
			cmds = append(cmds, loadArtifact(dir, c.Name, m.tab))
		case m.tab == tabSpecs:
			cmds = append(cmds, loadChangeDetail(m.client, c.Name))
		}
		return tea.Batch(cmds...)
	}
	if s, ok := m.selectedSpec(); ok {
		return loadSpecDetail(m.client, s.Name)
	}
	return nil
}

// refreshAll is the one real refresh path, shared by the `r` key, a completed
// streaming command, and a focus regain. It invalidates the CLI cache and every
// per-item cache — not just the list queries, which is why `r` used to leave the
// preview stale — then reloads the lists and force-reloads the previewed item.
//
// The entries backing the current render are carried over, so the visible
// preview keeps its content until fresh data lands instead of flashing
// "Loading…" (stale-while-revalidate). Everything else is dropped and reloads
// lazily when next selected, exactly as on first view.
func (m *Model) refreshAll() tea.Cmd {
	m.client.Invalidate()
	m.dropCaches()
	m.lastRefresh = time.Now()
	return tea.Batch(
		loadChanges(m.client),
		loadSpecs(m.client),
		loadArchived(m.rootPath),
		m.reloadPreview(),
	)
}

// dropCaches replaces every per-item cache with a fresh map, keeping only the
// entries that back the current render. Error entries for the current item are
// kept too, so a visible error message does not flicker back to "Loading…"
// before its reload lands; every other error is dropped, since a refresh is
// exactly when a previously failed load deserves another attempt.
func (m *Model) dropCaches() {
	m.statusCache = keepOne(m.statusCache, m.curChange)
	m.statusErr = keepOne(m.statusErr, m.curChange)
	m.changeDetail = keepOne(m.changeDetail, m.curChange)
	m.specsErr = keepOne(m.specsErr, m.curChange)
	m.archivedOv = keepOne(m.archivedOv, m.curChange)

	// Only the active tab's artifact is on screen; the other tabs reload lazily.
	key := ""
	if fileBackedTab(m.tab) {
		key = cacheKey(m.curChange, m.tab)
	}
	m.detailCache = keepOne(m.detailCache, key)
	m.detailErr = keepOne(m.detailErr, key)

	m.specErr = keepOne(m.specErr, m.curSpec)
	// m.specDetail holds the visible spec's requirements; its reload overwrites it.
}

// keepOne returns a fresh map holding only key's entry from src (empty when the
// key is absent or blank).
func keepOne[V any](src map[string]V, key string) map[string]V {
	out := make(map[string]V, 1)
	if key == "" {
		return out
	}
	if v, ok := src[key]; ok {
		out[key] = v
	}
	return out
}

// preserveSel restores panel p's selection to the row named name after its
// backing slice was swapped, so a refresh that reorders or drops rows does not
// silently move the cursor to a different item. When the item is gone the
// previous index is kept and clamped, landing on a neighbouring row. It runs
// beneath the list filter, via the visible-row accessors, so it composes with
// filter-relative indices.
func (m *Model) preserveSel(p panel, name string) {
	m.clampSel()
	if name == "" {
		return
	}
	if i := m.indexOfVisible(p, name); m.visibleNameAt(p, i) == name {
		m.sel[p] = i
	}
}

// handlePreviewKey routes keys while the preview pane is focused: focus toggle,
// search, match navigation, then item-specific tab/scroll handling.
func (m Model) handlePreviewKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if m.search.active() {
			m.clearSearch()
			m = m.refreshMain()
			return m, nil
		}
		m.activePane = paneNav
		m = m.refreshMain()
		return m, nil
	case "enter":
		m.activePane = paneNav
		m = m.refreshMain()
		return m, nil
	case "/":
		m.search = searchState{typing: true}
		m = m.refreshMain()
		return m, nil
	case "n":
		return m.jumpMatch(1), nil
	case "N":
		return m.jumpMatch(-1), nil
	}
	if _, ok := m.selectedChange(); ok {
		return m.previewChangeKey(msg)
	}
	if _, ok := m.selectedSpec(); ok {
		return m.previewSpecKey(msg)
	}
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

// previewChangeKey handles tab switching, task-cursor moves, and scrolling for a
// change preview.
func (m Model) previewChangeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "]", "right":
		m.cycleTab(1)
		return m.afterTabChange()
	case "[", "left":
		m.cycleTab(-1)
		return m.afterTabChange()
	case " ":
		return m.toggleTask()
	case "g":
		m.vp.GotoTop()
		return m, nil
	case "G":
		m.vp.GotoBottom()
		return m, nil
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
	m.vp, cmd = m.vp.Update(msg)
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

// cycleTab moves the change preview's artifact tab by delta, wrapping around.
// Pane-agnostic: both the list and the focused preview drive it with [ / ].
func (m *Model) cycleTab(delta int) {
	m.tab = artifactTab((int(m.tab) + delta + int(numTabs)) % int(numTabs))
}

// jumpRequirement scrolls the spec preview to the requirement delta steps away,
// wrapping around. Pane-agnostic, like cycleTab.
func (m *Model) jumpRequirement(delta int) {
	n := len(m.reqOffsets)
	if n == 0 {
		return
	}
	m.reqIdx = (m.reqIdx + delta + n) % n
	m.vp.SetYOffset(m.reqOffsets[m.reqIdx])
}

// previewSpecKey handles requirement jumping and scrolling for a spec preview.
func (m Model) previewSpecKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "]", "right":
		m.jumpRequirement(1)
		return m, nil
	case "[", "left":
		m.jumpRequirement(-1)
		return m, nil
	case "g":
		m.vp.GotoTop()
		return m, nil
	case "G":
		m.vp.GotoBottom()
		return m, nil
	}
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

// handleSearchInput edits the search query while typing is active. Every edit
// re-renders, which recomputes matches and jumps to the first one.
func (m Model) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		m.search.typing = false // confirm: keep the query and matches
	case tea.KeyEsc:
		m.clearSearch()
	case tea.KeyBackspace:
		if r := []rune(m.search.query); len(r) > 0 {
			m.search.query = string(r[:len(r)-1])
			m.search.idx = 0
		}
	case tea.KeySpace:
		m.search.query += " "
		m.search.idx = 0
	case tea.KeyRunes:
		m.search.query += string(msg.Runes)
		m.search.idx = 0
	default:
		return m, nil
	}
	m = m.refreshMain()
	return m, nil
}

// jumpMatch advances the current search match by delta (with wraparound) and
// re-renders so the new current match is highlighted and scrolled into view.
func (m Model) jumpMatch(delta int) Model {
	n := len(m.search.matches)
	if n == 0 {
		return m
	}
	m.search.idx = (m.search.idx + delta + n) % n
	return m.refreshMain()
}

// afterTabChange resets scroll/search and loads the newly selected tab's content
// if not cached.
func (m Model) afterTabChange() (tea.Model, tea.Cmd) {
	m.vp.GotoTop()
	m.taskCursor = 0
	m.clearSearch()
	cmd := m.ensurePreviewLoaded()
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

// panelLen is the number of rows a panel renders — the visible (filtered) count,
// which is what m.sel indexes.
func (m Model) panelLen(p panel) int {
	switch p {
	case panelChanges:
		return len(m.visibleChanges())
	case panelSpecs:
		return len(m.visibleSpecs())
	case panelArchive:
		return len(m.visibleArchived())
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
