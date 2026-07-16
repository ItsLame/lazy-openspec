## MODIFIED Requirements

### Requirement: Change preview
The application SHALL render, in the main pane, a live preview of the currently selected change as the selection moves, without requiring the user to press `enter`. The preview SHALL be a tabbed view whose tabs are `overview · proposal · specs · design · tasks`, opening on the `overview` tab, which shows the change's lifecycle state, task progress, and per-artifact completion status. The remaining tabs SHALL render the corresponding artifact content (proposal/design/tasks prose, and the change's spec deltas). The active tab SHALL persist as the selection moves from one change to another, rather than resetting to `overview`, so that a chosen artifact can be skimmed across successive changes.

#### Scenario: Preview on selection
- **WHEN** a change is selected in the Changes panel
- **THEN** the main pane immediately shows that change's `overview` tab — its status line and an artifact checklist indicating which artifacts are complete, in progress, or missing — with no `enter` required

#### Scenario: Preview updates as selection moves
- **WHEN** the selection moves from one change to another in the Changes panel
- **THEN** the main pane preview updates to the newly selected change's content

#### Scenario: Overview leads the tab bar
- **WHEN** a change preview is shown
- **THEN** the tab bar reads `overview · proposal · specs · design · tasks` with `overview` active by default

#### Scenario: Active tab persists across selection moves
- **WHEN** the `proposal` tab is active and the selection moves to another change
- **THEN** the preview shows the newly selected change's **proposal**, with `proposal` still the active tab, rather than resetting to `overview`

#### Scenario: Persisted tab with a missing artifact
- **WHEN** the `design` tab is active and the selection moves to a change that has no `design.md`
- **THEN** the preview shows the design tab's empty-state message for that change, and the tab bar still shows `design` as active

### Requirement: Drill into a change and switch artifacts
The application SHALL let the user switch the change preview's active artifact tab with `[` / `]` from **either** pane: while the list holds focus (without pressing `enter` first, leaving focus on the list) and while the preview pane holds focus. Pressing `enter` on a selected change SHALL move keyboard focus into the preview pane (rather than opening a separate screen), where `[` / `]` (or left/right) also switch tabs, and `esc` SHALL return focus to the list. The list selection and the two-pane layout SHALL remain visible throughout.

#### Scenario: Switch tabs from the list
- **WHEN** a change is selected in the Changes panel, the list holds focus, and the user presses `]`
- **THEN** the preview's active tab advances and its content is rendered, while keyboard focus remains on the list so `j` / `k` still move the selection

#### Scenario: Focus the preview with enter
- **WHEN** the user presses `enter` on a change while the list is focused
- **THEN** keyboard focus moves to the preview pane, the pane's border highlight indicates it is active, and scroll/search/tab keys are routed to it

#### Scenario: Switch to the design tab
- **WHEN** the preview pane is focused for a change and the user presses `]` until the design tab is active
- **THEN** the main pane renders the change's `design.md`, or an empty-state message if that artifact does not exist yet

#### Scenario: Return focus to the list
- **WHEN** the preview pane is focused and the user presses `esc` with no active search
- **THEN** focus returns to the list with the previously selected change still highlighted

### Requirement: Archive browsing
The Archive panel SHALL list archived changes and preview their artifacts read-only, sourced entirely from disk. The application SHALL NOT invoke `openspec status` or `openspec show` for archived changes; instead the `overview` tab SHALL be derived from the on-disk artifact files (task counts and artifact presence) and the `specs` tab SHALL show a note that the deltas were merged into the main specs on archive. The preview SHALL reach a rendered state and SHALL NOT remain on a loading placeholder. Tab switching with `[` / `]` SHALL work for archived changes from the list as well as from the focused preview.

#### Scenario: Preview an archived change without hanging
- **WHEN** the user selects an archived change in the Archive panel
- **THEN** the main pane renders its overview and artifact content from disk, and never remains on a "Loading" placeholder

#### Scenario: Archived specs tab explains the merge
- **WHEN** the preview's `specs` tab is shown for an archived change
- **THEN** the pane shows a note that the change's spec deltas were merged into the main specs when it was archived, rather than a loading placeholder or error

#### Scenario: Archived preview is read-only
- **WHEN** the preview pane is focused for an archived change on the tasks tab
- **THEN** task toggling is disabled and the pane indicates the change is archived (read-only)

#### Scenario: Switch archived tabs from the list
- **WHEN** an archived change is selected, the Archive panel holds focus, and the user presses `]`
- **THEN** the preview's active tab advances and renders that archived change's artifact from disk, with focus remaining on the list
