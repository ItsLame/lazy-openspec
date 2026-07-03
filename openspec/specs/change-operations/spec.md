# change-operations Specification

## Purpose
TBD - created by archiving change add-lazy-openspec-tui. Update Purpose after archive.
## Requirements
### Requirement: Toggle task completion
The application SHALL let the user toggle the completion state of the selected task in the tasks view (`space`), persisting the change by rewriting the corresponding `- [ ]` / `- [x]` marker in the change's `tasks.md`, and SHALL update the displayed progress immediately.

#### Scenario: Check a task
- **WHEN** an unchecked task is selected and the user presses `space`
- **THEN** the task's marker in `tasks.md` becomes `- [x]`, the task renders as completed, and the group and change progress indicators update

#### Scenario: Uncheck a task
- **WHEN** a checked task is selected and the user presses `space`
- **THEN** the task's marker in `tasks.md` becomes `- [ ]` and the progress indicators update accordingly

#### Scenario: Toggle preserves surrounding content
- **WHEN** a task is toggled
- **THEN** only that task's checkbox marker changes and all other lines, task text, numbering, and formatting in `tasks.md` are preserved byte-for-byte

### Requirement: Run workflow commands
The application SHALL let the user run `openspec validate`, `openspec apply`, and `openspec archive` for the selected change from within the TUI, streaming the command's output into the command-log pane and reflecting success or failure.

#### Scenario: Validate a change
- **WHEN** the user triggers validate (`v`) on the selected change
- **THEN** the application runs `openspec validate <name>`, streams its output into the command log, and shows a success or failure indicator when it completes

#### Scenario: Command failure surfaced
- **WHEN** a triggered command exits with a non-zero status
- **THEN** the command log shows the error output and the app indicates failure without crashing or leaving the UI in an inconsistent state

### Requirement: Confirm destructive actions
The application SHALL require an explicit confirmation before running an action that mutates the OpenSpec tree irreversibly, specifically `openspec archive`.

#### Scenario: Confirm before archiving
- **WHEN** the user triggers archive (`A`) on a change
- **THEN** a confirmation prompt is shown, and the archive command runs only if the user confirms; cancelling leaves the change unchanged

### Requirement: Refresh after mutation
The application SHALL refresh the affected data after a task toggle or a workflow command completes, so the panels reflect the new state without requiring a manual refresh.

#### Scenario: Auto-refresh after archive
- **WHEN** an archive command completes successfully
- **THEN** the archived change moves out of the active/completed groups and appears under the Archive panel without the user pressing refresh

