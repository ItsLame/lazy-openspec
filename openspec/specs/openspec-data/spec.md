# openspec-data Specification

## Purpose
TBD - created by archiving change add-lazy-openspec-tui. Update Purpose after archive.
## Requirements
### Requirement: Load data from the OpenSpec CLI
The application SHALL obtain all changes, specs, and status data by invoking the installed `openspec` CLI with its `--json` flag (`list`, `list --specs`, `show`, `status`, `instructions`) and parsing the structured output, rather than parsing the `openspec/` markdown tree directly.

#### Scenario: Load the change list
- **WHEN** the application starts and needs the list of changes
- **THEN** it runs `openspec list --json`, parses the result, and populates the Changes panel with the returned entries

#### Scenario: Load change status
- **WHEN** a change is selected
- **THEN** the application runs `openspec status --change <name> --json` and uses the parsed artifact statuses and task progress to render the change's state

### Requirement: Caching and manual refresh
The application SHALL cache CLI results in memory to avoid spawning a subprocess on every keystroke, and SHALL expose a manual refresh action (`r`) that re-runs the relevant CLI queries and updates the view. A refresh (manual, post-command, or focus-driven) SHALL invalidate the per-item caches — status, artifact content, spec deltas, spec details, and archived overviews — not only the list queries, so previously previewed items re-render from fresh data. During a refresh the currently visible preview SHALL keep displaying its previous content until fresh data arrives rather than flashing a loading placeholder, the preview's scroll position SHALL be preserved, and each list panel's selection SHALL be preserved by item identity when the item still exists after the reload.

#### Scenario: Selection reuses cache
- **WHEN** the user moves the selection between two already-loaded changes
- **THEN** the previews render from cached data without spawning a new `openspec` process

#### Scenario: Manual refresh
- **WHEN** the user presses `r`
- **THEN** the application re-runs the queries backing the current view and updates the panels with the fresh results

#### Scenario: Refresh updates previewed content
- **WHEN** a previewed artifact (e.g. the tasks tab of a change) is modified on disk outside the TUI and a refresh occurs
- **THEN** the preview re-renders with the modified content instead of the cached pre-refresh version

#### Scenario: Refresh preserves selection and scroll
- **WHEN** a refresh reloads the lists while the user has a change selected and its preview scrolled partway down
- **THEN** the same change remains selected (found by name in the reloaded list) and the preview's scroll offset is unchanged if the content still accommodates it

#### Scenario: Refresh removes a vanished item gracefully
- **WHEN** the selected change no longer exists after a refresh (e.g. it was archived externally)
- **THEN** the selection falls back to a valid neighboring row and the preview follows the new selection without an error

### Requirement: Graceful degradation
The application SHALL detect when the `openspec` CLI is missing or when no `openspec/` root can be resolved, and SHALL present a clear, non-crashing message instead of failing with an unhandled error.

#### Scenario: CLI not installed
- **WHEN** the `openspec` executable cannot be found on `PATH`
- **THEN** the application displays an actionable message explaining that `openspec` must be installed and exits without a stack trace

#### Scenario: No OpenSpec root
- **WHEN** the application is launched in a directory with no resolvable `openspec/` root
- **THEN** it displays a message indicating no OpenSpec project was found rather than rendering empty panels silently

### Requirement: Store selection
The application SHALL support targeting a registered OpenSpec store, passing `--store <id>` through to the CLI commands that accept it when a store is selected.

#### Scenario: Launch against a store
- **WHEN** the user launches the application with a store id (e.g. `lazy-openspec --store <id>`)
- **THEN** every CLI query that supports `--store` is invoked with that store id and the panels reflect that store's changes and specs

### Requirement: Tolerant decoding of CLI identifiers
The application SHALL decode the spec identifier from the `openspec` CLI's JSON output tolerantly, accepting the value whether the CLI provides it under the field `id` or the field `name`, so that spec identities populate correctly across CLI versions. A decoded spec SHALL never carry an empty identifier when the CLI supplied one under either field.

#### Scenario: CLI reports specs keyed by id
- **WHEN** `openspec list --specs --json` returns spec entries using the field `id` (as in openspec 1.5.0)
- **THEN** each decoded spec's identifier is populated from that `id`, and it is this identifier that is passed to subsequent commands such as `openspec spec show <id> --json`

#### Scenario: CLI reports specs keyed by name
- **WHEN** a CLI version returns spec entries using the field `name` instead of `id`
- **THEN** the decoded spec's identifier is populated from that `name` without code changes, and the application behaves identically

### Requirement: Focus-driven refresh
The application SHALL enable terminal focus reporting and SHALL automatically perform a full data refresh when the terminal regains focus after having been blurred, so that returning to the TUI from another pane, tab, or window presents current data without a manual refresh. The application SHALL NOT poll or refresh while blurred. Focus-driven refreshes SHALL be debounced and SHALL be skipped while a streaming workflow command is running. On terminals that do not report focus events, the application SHALL behave exactly as it does today, with manual refresh as the fallback.

#### Scenario: Refresh on regaining focus
- **WHEN** the `openspec/` tree is modified externally while the terminal is blurred (e.g. an agent edits artifacts from another tmux pane) and the terminal then regains focus
- **THEN** the application refreshes the list panels and the currently previewed item, and the view reflects the external modifications

#### Scenario: Stale while blurred
- **WHEN** the terminal is blurred and external modifications occur
- **THEN** the application performs no refresh and spawns no subprocesses until focus returns

#### Scenario: Focus flapping is debounced
- **WHEN** the terminal loses and regains focus repeatedly within the debounce window
- **THEN** at most one refresh is performed for the burst

#### Scenario: No refresh while a command is running
- **WHEN** the terminal regains focus while a streaming `openspec` command started from the TUI is still running
- **THEN** the focus-driven refresh is skipped, and the command's own completion refresh brings the view up to date

#### Scenario: Terminal without focus reporting
- **WHEN** the application runs in a terminal that does not emit focus events (e.g. tmux without `focus-events on`)
- **THEN** no focus-driven refresh occurs, no errors are shown, and manual refresh (`r`) remains available and effective

