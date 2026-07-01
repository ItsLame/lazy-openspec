## Why

OpenSpec's only visual surface today is `openspec view` — a one-shot, non-interactive printout. Reading and steering a change means juggling `list`, `show`, `status`, `validate`, and `apply` across separate invocations, and reading raw markdown artifacts in an editor. There is no single, keyboard-driven place to browse changes and specs, read artifacts comfortably, tick off tasks, and run the workflow commands. `lazy-openspec` fills that gap with a `lazygit`/`lazydocker`-style terminal UI so the whole spec-driven loop lives in one screen.

## What Changes

- Add a new standalone terminal application, **`lazy-openspec`** (Go + Bubble Tea + Lip Gloss + Glamour), that ships as a single static binary.
- Provide a **lazygit-style multi-panel shell**: stacked, numbered panels on the left (Changes, Specs, Archive), a main content pane on the right, and a command-log/status bar at the bottom, navigated by tab and number keys.
- **Browse changes** grouped by lifecycle (draft / active / completed) with progress bars, and drill into a change's artifacts via a `proposal · specs · design · tasks` tab bar.
- **Browse specs** with requirements and their WHEN/THEN scenarios rendered semantically.
- **Render artifacts to be easily readable**: Glamour for prose sections, plus semantic rendering of OpenSpec's known structures (requirements, scenarios, WHEN/THEN, task checklists) instead of raw markdown.
- **Interactively toggle tasks**: check/uncheck items in `tasks.md` from the tasks pane, persisted back to disk.
- **Run workflow actions** from the TUI — `openspec validate`, `apply`, and `archive` — streaming their output into the command-log pane and refreshing afterward.
- **Source all data by shelling out to `openspec … --json`** (`list`, `show`, `status`, `instructions`), keeping the TUI decoupled from OpenSpec internals.

## Capabilities

### New Capabilities
- `tui-shell`: The lazygit-style application shell — panel layout, focus cycling, global keybindings, help overlay, and the command-log/status bar.
- `openspec-data`: The data-access layer that invokes `openspec … --json`, parses and caches results, refreshes on demand, and degrades gracefully when the CLI or an `openspec/` root is missing.
- `change-navigation`: The changes dashboard (lifecycle grouping, progress) and change drill-in with artifact tab switching and viewing.
- `spec-navigation`: The specs list and the requirement/scenario detail view.
- `artifact-rendering`: Readable rendering of artifacts — Glamour for prose plus semantic rendering of OpenSpec blocks (requirements, scenarios, WHEN/THEN, task checklists), with word-wrapping and theming.
- `change-operations`: Write actions on a change — toggling task checkboxes (persisted to `tasks.md`) and running `openspec validate`/`apply`/`archive` with streamed output.

### Modified Capabilities
<!-- None. This is a greenfield repository with no existing specs. -->

## Impact

- **New codebase**: a Go module in this repository (`lazy-openspec`) producing the `lazy-openspec` binary. New folders for the TUI (Bubble Tea models/views), the OpenSpec data client, and the rendering layer.
- **New dependencies**: `bubbletea`, `lipgloss`, `glamour` (Charmbracelet), plus a terminal backend.
- **External dependency (runtime)**: relies on the `openspec` CLI being installed and on its `--json` output contract (`list`, `show`, `status`, `instructions`). Version drift in that contract is the main integration risk.
- **Filesystem writes**: task toggling edits `tasks.md` in place; workflow actions invoke `openspec` which mutates the `openspec/` tree (e.g. `archive`).
- **No changes to OpenSpec itself**: this is an additive, standalone tool; the existing `openspec view` command is untouched.
