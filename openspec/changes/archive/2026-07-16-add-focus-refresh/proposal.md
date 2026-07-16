# Add Focus Refresh

## Why

The TUI is a snapshot: data is fetched once at startup, every preview is lazily loaded into in-memory caches, and nothing invalidates those caches when the `openspec/` tree changes on disk. In the workflow this tool is built for — an AI agent or a CLI in an adjacent tmux pane/tab writing artifacts while lazy-openspec sits open — the dashboard silently drifts out of date. Worse, the manual `r` refresh only reloads the three list panels; per-item caches (status, proposal/design/tasks content, spec deltas, spec details) are never cleared, so even a deliberate refresh leaves the preview stale. `lazygit` solves this with terminal focus events: it refetches when the terminal pane regains focus. Bubble Tea ships the same primitive (`tea.WithReportFocus()`, `tea.FocusMsg`/`tea.BlurMsg`), and the bubbletea version already in `go.mod` (v1.3.10) supports it — no new dependencies.

## What Changes

- **Refresh on focus regain.** Enable terminal focus reporting and, when the terminal regains focus after having been blurred, run a full data refresh automatically — the multi-pane/multi-tab "switch back and it's current" behavior of `lazygit`. While blurred, the view is allowed to be stale (no background polling, no file watching).
- **One real refresh path.** Introduce a single full-refresh routine that invalidates the CLI client cache *and* the per-item caches, then reloads the lists and the currently previewed item. Manual `r`, post-command refresh, and focus-regain all route through it. This fixes the existing bug where `r` leaves preview content stale.
- **Unobtrusive refresh.** The visible preview refreshes stale-while-revalidate: old content stays on screen until fresh data arrives (no "Loading…" flash), scroll position is kept, and the list selection is preserved by item name rather than index so a background refresh never yanks the cursor to a different row.
- **Guarded triggers.** Focus-regain refreshes are debounced (rapid focus flapping does not stack subprocess spawns) and skipped while a streaming `openspec` command is running (its completion already triggers a refresh).
- **Graceful fallback.** Terminals without focus reporting (or tmux without `focus-events on`) simply never emit the events; behavior degrades to today's manual `r`, which now actually works.

## Capabilities

### New Capabilities

None — data freshness already lives in `openspec-data`.

### Modified Capabilities

- `openspec-data`: The "Caching and manual refresh" requirement is amended — refresh must invalidate per-item caches (not just re-run list queries) and preserve selection and scroll position. A new "Focus-driven refresh" requirement is added: refresh automatically on terminal focus regain, debounced, skipped while a command is running, with stale-while-blurred semantics and graceful degradation on terminals without focus reporting.

## Impact

- **Code:** `cmd/lazy-openspec/main.go` (add `tea.WithReportFocus()`), `internal/tui/model.go` (focus + last-refresh state), `internal/tui/update.go` (handle `tea.FocusMsg`/`tea.BlurMsg`; extract the shared `refreshAll` path; selection preservation by name), `internal/tui/commands.go` (no new loaders expected).
- **Dependencies:** none — bubbletea v1.3.10 already provides focus reporting.
- **Terminal support:** focus reporting requires DEC mode 1004 (iTerm2, kitty, WezTerm, Alacritty, Ghostty, Windows Terminal). tmux users need `set -g focus-events on`; document this in the README.
- **In-flight changes:** `add-live-preview-panes` and `refine-list-panels` both touch `model.go`/`update.go`. This change is behaviorally orthogonal (data lifecycle vs. interaction) but will conflict textually; implement it after those land. Selection preservation must compose with `refine-list-panels`' filter-relative indices.
