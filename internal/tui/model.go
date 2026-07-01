package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/itslame/lazy-openspec/internal/openspec"
	"github.com/itslame/lazy-openspec/internal/render"
)

type screen int

const (
	screenDashboard screen = iota
	screenChangeDetail
	screenSpecDetail
)

type panel int

const (
	panelChanges panel = iota
	panelSpecs
	panelArchive
	numPanels
)

var panelTitles = [numPanels]string{"1 Changes", "2 Specs", "3 Archive"}

type artifactTab int

const (
	tabProposal artifactTab = iota
	tabSpecs
	tabDesign
	tabTasks
	numTabs
)

var tabNames = [numTabs]string{"proposal", "specs", "design", "tasks"}

// confirmState drives a yes/no prompt for destructive actions.
type confirmState struct {
	prompt string
	onYes  func(*Model) tea.Cmd
}

// minCols/minRows are the smallest usable terminal size.
const (
	minCols = 60
	minRows = 18
)

// Model is the root Bubble Tea model for lazy-openspec.
type Model struct {
	client *openspec.Client
	md     *render.Markdown
	sem    *render.Semantic

	width, height int
	ready         bool

	screen screen
	focus  panel

	rootPath string
	loadErr  error

	changes  []openspec.ChangeSummary
	specs    []openspec.SpecSummary
	archived []openspec.ChangeSummary

	sel [numPanels]int

	statusCache map[string]openspec.Status

	// change detail state
	curChange    string
	curChangeDir string
	curArchived  bool
	tab          artifactTab
	taskCursor   int                              // selected task index in the tasks tab
	detailCache  map[string]string                // "change/tab" -> markdown
	changeDetail map[string]openspec.ChangeDetail // change -> deltas

	// spec detail state
	curSpec    string
	specDetail *openspec.SpecDetail
	reqIdx     int
	reqOffsets []int // vp line offset of each requirement, for n/p navigation

	// actions overlay
	showActions bool

	vp viewport.Model

	logs    []string
	cmdCh   chan tea.Msg
	running bool

	showHelp bool
	confirm  *confirmState

	quitting bool
}

// New builds the root model bound to an openspec client.
func New(client *openspec.Client) Model {
	return Model{
		client:       client,
		sem:          render.NewSemantic(),
		md:           render.NewMarkdown(80),
		statusCache:  map[string]openspec.Status{},
		detailCache:  map[string]string{},
		changeDetail: map[string]openspec.ChangeDetail{},
		cmdCh:        make(chan tea.Msg, 128),
		vp:           viewport.New(80, 20),
	}
}

// Init kicks off the initial data loads.
func (m Model) Init() tea.Cmd {
	return tea.Batch(loadChanges(m.client), loadSpecs(m.client))
}

// ---- layout -----------------------------------------------------------------

type dims struct {
	leftW, mainW int
	bodyH        int
	logH, hintH  int
	panelH       [numPanels]int
	vpW, vpH     int
}

// layout computes region sizes from the current terminal dimensions.
func (m Model) layout() dims {
	d := dims{hintH: 1, logH: 6}
	d.bodyH = m.height - d.hintH - d.logH
	if d.bodyH < 6 {
		d.bodyH = 6
	}
	d.leftW = m.width / 3
	if d.leftW < 26 {
		d.leftW = 26
	}
	if d.leftW > 40 {
		d.leftW = 40
	}
	d.mainW = m.width - d.leftW
	base := d.bodyH / int(numPanels)
	for i := range d.panelH {
		d.panelH[i] = base
	}
	d.panelH[numPanels-1] += d.bodyH - base*int(numPanels)
	d.vpW = d.mainW - 4
	if d.vpW < 10 {
		d.vpW = 10
	}
	// Reserve lines inside the main box: 2 border + title + subtitle + scroll.
	d.vpH = d.bodyH - 5
	if d.vpH < 3 {
		d.vpH = 3
	}
	return d
}

// selectedChange returns the currently highlighted change summary, if any.
func (m Model) selectedChange() (openspec.ChangeSummary, bool) {
	switch m.focus {
	case panelChanges:
		if i := m.sel[panelChanges]; i >= 0 && i < len(m.changes) {
			return m.changes[i], true
		}
	case panelArchive:
		if i := m.sel[panelArchive]; i >= 0 && i < len(m.archived) {
			return m.archived[i], true
		}
	}
	return openspec.ChangeSummary{}, false
}

// selectedSpec returns the currently highlighted spec, if any.
func (m Model) selectedSpec() (openspec.SpecSummary, bool) {
	if m.focus == panelSpecs {
		if i := m.sel[panelSpecs]; i >= 0 && i < len(m.specs) {
			return m.specs[i], true
		}
	}
	return openspec.SpecSummary{}, false
}
