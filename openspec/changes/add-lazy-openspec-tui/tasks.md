## 1. Project scaffolding

- [x] 1.1 Initialize the Go module (`go mod init`) and commit a `go.mod`/`go.toml` layout with `cmd/lazy-openspec` as the binary entrypoint
- [x] 1.2 Add dependencies: `bubbletea`, `lipgloss`, `glamour`, and a terminal color/backend helper
- [x] 1.3 Create the package layout: `internal/openspec` (data client), `internal/render` (rendering), `internal/tui` (Bubble Tea models)
- [x] 1.4 Wire a minimal `main` that starts a Bubble Tea program, enters the alternate screen, and quits cleanly on `q`/`ctrl+c`
- [x] 1.5 Add a build/release setup (`Makefile` or `goreleaser` config) and a README stub

## 2. OpenSpec data client (openspec-data)

- [x] 2.1 Implement a subprocess runner that executes `openspec … --json`, captures stdout/stderr, and decodes tolerantly (ignore unknown fields)
- [x] 2.2 Add typed models + calls for `list --json`, `list --specs --json`, `status --change <name> --json`, and `show --json`
- [x] 2.3 Thread an optional `--store <id>` flag through every call that supports it
- [x] 2.4 Add an in-memory cache keyed by (command, args) with an explicit invalidate/refresh entrypoint
- [x] 2.5 Detect a missing `openspec` binary and a missing/unresolvable `openspec/` root, returning typed errors (not panics)
- [x] 2.6 Expose all CLI calls as async operations suitable for Bubble Tea `Cmd`s

## 3. TUI shell (tui-shell)

- [x] 3.1 Build the root model holding focus state and the three left panels (Changes, Specs, Archive) plus the main pane
- [x] 3.2 Compose the layout with Lip Gloss (stacked bordered panels, main pane, bottom bar) and reflow on `WindowSizeMsg`
- [x] 3.3 Implement focus: `tab`/`shift+tab` cycling, number keys `1`/`2`/`3` to jump, and focused-panel border highlight
- [x] 3.4 Implement in-panel selection movement (`↑`/`↓`, `j`/`k`) that updates the main-pane preview
- [x] 3.5 Add the bottom bar with a context-sensitive keybinding hint line and a command-log region
- [x] 3.6 Add the `?` help overlay listing current-context keybindings, dismissible with `?`/`esc`/`q`
- [x] 3.7 Add a minimum-terminal-size guard and graceful-degradation messages (no CLI / no root)

## 4. Change navigation (change-navigation)

- [x] 4.1 Populate the Changes panel grouped by lifecycle (draft/active/completed) with status glyphs
- [x] 4.2 Render progress bars + percentages for active changes from task counts
- [x] 4.3 Render the main-pane change preview: status line + per-artifact completion checklist
- [x] 4.4 Implement `enter` to open the change detail view and `esc` to return with selection preserved
- [x] 4.5 Add the artifact tab bar (`proposal · specs · design · tasks`) with `[`/`]` and `←`/`→` switching, including empty-artifact states
- [x] 4.6 Populate the Archive panel and allow read-only viewing of archived changes' artifacts
- [x] 4.7 Add empty-state messages for the Changes and Archive panels

## 5. Spec navigation (spec-navigation)

- [x] 5.1 Populate the Specs panel with each capability name and requirement count, stably sorted
- [x] 5.2 Implement the spec detail view listing requirements with descriptions and their scenarios
- [x] 5.3 Add `n`/`p` navigation between requirements and scrolling of long content
- [x] 5.4 Add an empty-state message when the root has no specs

## 6. Artifact rendering (artifact-rendering)

- [x] 6.1 Configure Glamour with a theme selected from terminal background / `NO_COLOR`, and render prose sections wrapped to pane width
- [x] 6.2 Re-wrap rendered markdown on pane width changes without breaking glyphs
- [x] 6.3 Implement semantic rendering of requirements (badged headers) and scenarios with aligned/colored `WHEN`/`THEN`
- [x] 6.4 Implement semantic rendering of `tasks.md` as grouped checklists with per-group progress and `☐`/`✔` glyphs, dimming completed tasks
- [x] 6.5 Implement a monochrome fallback style that stays readable without color

## 7. Task editing & change operations (change-operations)

- [x] 7.1 Implement `space` to toggle the selected task, rewriting only its `- [ ]`/`- [x]` marker in `tasks.md`
- [x] 7.2 Match the target task by number + text (re-reading the file first); abort with a "changed on disk, refresh" message if the marker isn't found
- [x] 7.3 Update group and change progress indicators immediately after a toggle
- [x] 7.4 Add actions to run `openspec validate` (`v`) and `openspec apply` (`a`), streaming output into the command-log pane with success/failure indication
- [x] 7.5 Add `openspec archive` (`A`) behind a confirmation prompt, and an `x` context menu listing available actions for the selection
- [x] 7.6 Auto-refresh the affected data after a toggle or command completes; surface non-zero exits without leaving the UI inconsistent
- [x] 7.7 Detect commands that require interactivity and report that the user should run them in a shell instead

## 8. Polish, resilience & release

- [x] 8.1 Add unit tests for the data client (JSON decoding, store flag, error cases) and the task-toggle file rewrite (byte-preservation)
- [x] 8.2 Add rendering tests/snapshots for the semantic renderers (scenarios, task checklists)
- [x] 8.3 Manually verify the full loop against a sample `openspec/` project (browse → toggle → validate → apply → archive)
- [x] 8.4 Write user-facing README (install, keybindings, screenshots/gif) and document the tested OpenSpec CLI version(s)
- [x] 8.5 Produce release binaries (and/or `go install` + Homebrew instructions)
