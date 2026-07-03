# change-navigation Specification

## Purpose
TBD - created by archiving change add-lazy-openspec-tui. Update Purpose after archive.
## Requirements
### Requirement: Changes grouped by lifecycle
The Changes panel SHALL list changes grouped by lifecycle state — draft (no tasks), active (some tasks done), and completed (all tasks done) — and SHALL show a progress indicator and completion percentage for active changes.

#### Scenario: Grouped listing
- **WHEN** the Changes panel renders with a mix of draft, active, and completed changes
- **THEN** each change appears under its correct lifecycle group with a status glyph, and active changes display a progress bar and percentage derived from their task counts

#### Scenario: Empty state
- **WHEN** there are no changes in the resolved OpenSpec root
- **THEN** the Changes panel shows an explicit empty-state message rather than an empty box

### Requirement: Change preview
The application SHALL show, in the main pane, a summary of the currently selected change including its lifecycle state, task progress, and per-artifact completion status (proposal, specs, design, tasks).

#### Scenario: Preview on selection
- **WHEN** a change is selected in the Changes panel
- **THEN** the main pane shows the change's status line and an artifact checklist indicating which artifacts are complete, in progress, or missing

### Requirement: Drill into a change and switch artifacts
The application SHALL let the user open a selected change (`enter`) into a detail view with a `proposal · specs · design · tasks` tab bar, switch the active artifact tab (`[` / `]` or left/right), and return to the dashboard (`esc`).

#### Scenario: Open change and view proposal
- **WHEN** the user presses `enter` on a change
- **THEN** the detail view opens with the proposal tab active and the rendered proposal shown in the main pane

#### Scenario: Switch to the design tab
- **WHEN** the change detail view is open and the user presses `]` until the design tab is active
- **THEN** the main pane renders the change's `design.md`, or an empty-state message if that artifact does not exist yet

#### Scenario: Return to dashboard
- **WHEN** the user presses `esc` in the change detail view
- **THEN** the application returns to the dashboard with the previously selected change still highlighted

### Requirement: Archive browsing
The Archive panel SHALL list archived changes and allow viewing their artifacts read-only.

#### Scenario: Open an archived change
- **WHEN** the user selects an archived change and presses `enter`
- **THEN** its artifacts are shown in the detail view in the same read-only manner as active changes

