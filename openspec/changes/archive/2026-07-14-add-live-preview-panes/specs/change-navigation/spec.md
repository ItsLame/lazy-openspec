## MODIFIED Requirements

### Requirement: Change preview
The application SHALL render, in the main pane, a live preview of the currently selected change as the selection moves, without requiring the user to press `enter`. The preview SHALL be a tabbed view whose tabs are `overview · proposal · specs · design · tasks`, defaulting to the `overview` tab, which shows the change's lifecycle state, task progress, and per-artifact completion status. The remaining tabs SHALL render the corresponding artifact content (proposal/design/tasks prose, and the change's spec deltas).

#### Scenario: Preview on selection
- **WHEN** a change is selected in the Changes panel
- **THEN** the main pane immediately shows that change's `overview` tab — its status line and an artifact checklist indicating which artifacts are complete, in progress, or missing — with no `enter` required

#### Scenario: Preview updates as selection moves
- **WHEN** the selection moves from one change to another in the Changes panel
- **THEN** the main pane preview updates to the newly selected change's content

#### Scenario: Overview leads the tab bar
- **WHEN** a change preview is shown
- **THEN** the tab bar reads `overview · proposal · specs · design · tasks` with `overview` active by default

### Requirement: Drill into a change and switch artifacts
The application SHALL let the user press `enter` on a selected change to move keyboard focus into the preview pane (rather than opening a separate screen), switch the active artifact tab there with `[` / `]` (or left/right), and return focus to the list with `esc`. The list selection and the two-pane layout SHALL remain visible throughout.

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
The Archive panel SHALL list archived changes and preview their artifacts read-only, sourced entirely from disk. The application SHALL NOT invoke `openspec status` or `openspec show` for archived changes; instead the `overview` tab SHALL be derived from the on-disk artifact files (task counts and artifact presence) and the `specs` tab SHALL show a note that the deltas were merged into the main specs on archive. The preview SHALL reach a rendered state and SHALL NOT remain on a loading placeholder.

#### Scenario: Preview an archived change without hanging
- **WHEN** the user selects an archived change in the Archive panel
- **THEN** the main pane renders its overview and artifact content from disk, and never remains on a "Loading" placeholder

#### Scenario: Archived specs tab explains the merge
- **WHEN** the preview's `specs` tab is shown for an archived change
- **THEN** the pane shows a note that the change's spec deltas were merged into the main specs when it was archived, rather than a loading placeholder or error

#### Scenario: Archived preview is read-only
- **WHEN** the preview pane is focused for an archived change on the tasks tab
- **THEN** task toggling is disabled and the pane indicates the change is archived (read-only)
