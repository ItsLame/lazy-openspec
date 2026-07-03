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
The application SHALL cache CLI results in memory to avoid spawning a subprocess on every keystroke, and SHALL expose a manual refresh action (`r`) that re-runs the relevant CLI queries and updates the view.

#### Scenario: Selection reuses cache
- **WHEN** the user moves the selection between two already-loaded changes
- **THEN** the previews render from cached data without spawning a new `openspec` process

#### Scenario: Manual refresh
- **WHEN** the user presses `r`
- **THEN** the application re-runs the queries backing the current view and updates the panels with the fresh results

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

