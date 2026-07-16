# Design: Add Focus Refresh

## Context

The TUI loads data exactly three ways today: `Init` fires `loadChanges` + `loadSpecs` once at startup; `ensurePreviewLoaded` lazily dispatches a loader when a cache entry is missing; and two mutation paths (`r` key, `cmdDoneMsg` after a streaming command) call `client.Invalidate()` and reload the three list panels. All per-item data lives in in-memory maps on the model — `statusCache`, `detailCache` (artifact markdown), `changeDetail` (spec deltas), `archivedOv`, `specDetail` — and **no refresh path clears them**: `ensurePreviewLoaded` only dispatches on a cache miss, so after `r` the lists are fresh but every preview renders stale content until the process restarts. There is no filesystem watching and no terminal focus reporting; `cmd/lazy-openspec/main.go` constructs the program with `tea.WithAltScreen()` only.

The target workflow is lazygit-style: the user keeps lazy-openspec open in one tmux pane/tab while an AI agent or the `openspec` CLI mutates the `openspec/` tree from another. When the user switches back, the dashboard should be current.

Constraints:

- bubbletea v1.3.10 (already in `go.mod`) provides `tea.WithReportFocus()`, `tea.FocusMsg`, and `tea.BlurMsg`. No dependency changes.
- Two in-flight changes (`add-live-preview-panes`, `refine-list-panels`) rewrite parts of `model.go`/`update.go`. This change lands after them.

## Goals / Non-Goals

**Goals:**

- Automatically refresh all data when the terminal regains focus after having been blurred.
- Consolidate `r`, post-command refresh, and focus-regain into one `refreshAll` routine that actually invalidates per-item caches — fixing the stale-preview bug in today's `r`.
- Keep refresh unobtrusive: no "Loading…" flash for the visible preview, scroll position kept, selection preserved by name.
- Degrade gracefully on terminals without focus reporting.

**Non-Goals:**

- No filesystem watching (fsnotify) and no periodic polling ticker. Focus events cover the stated workflow without a new dependency or background subprocess churn; either can be added later if a real gap shows up.
- No refreshing while blurred — stale-while-blurred is explicitly acceptable.
- No attempt to detect focus in terminals that lack DEC mode 1004; the manual `r` remains the fallback.

## Decisions

### 1. Terminal focus reporting, not file watching or polling

Add `tea.WithReportFocus()` to the `tea.NewProgram` options. Bubble Tea then delivers `tea.FocusMsg`/`tea.BlurMsg` when the terminal (or tmux pane, with `focus-events on`) gains/loses focus.

*Alternatives considered:* **fsnotify** watches would catch every write but add a dependency, per-platform edge cases (macOS FSEvents coalescing, watch-descriptor limits on Linux), and a debounce problem of their own — agents write many files in bursts. **A polling ticker** (lazygit also polls every ~10s) spawns `openspec` subprocesses while nobody is looking, which is exactly the waste the cache exists to avoid. Focus events fire precisely when freshness starts to matter and cost nothing while idle. They also match the user's stated model: "stale while unblurred is fine, fresh when I come back."

### 2. One `refreshAll` routine with stale-while-revalidate for the visible preview

Extract a `refreshAll()` method used by `r`, `cmdDoneMsg`, and focus-regain:

1. `client.Invalidate()`.
2. Replace every per-item cache map with a fresh one, **except** the entries backing the current render (current change's status / active-tab artifact / spec deltas / archived overview, or current spec's detail). Those are carried over so the on-screen preview keeps rendering old content instead of flashing "Loading…".
3. Batch `loadChanges`, `loadSpecs`, `loadArchived`, plus a `reloadPreview()` that dispatches the loaders for the current selection and active tab **unconditionally** (a force-load variant of `ensurePreviewLoaded`'s cache-miss checks).

Fresh messages overwrite the carried-over entries and `refreshMain` re-renders; the viewport keeps its `YOffset` (bubbles' viewport clamps if the content shrank), so an unchanged artifact re-renders imperceptibly. Non-visible items were dropped in step 2, so they reload through the existing lazy "Loading…" path when next selected — same UX as first view.

*Alternative considered:* a generation counter on cache entries (keep everything, mark stale, re-dispatch on read) is more machinery for the same observable behavior; wholesale replacement with a carve-out for the visible keys is a few lines.

### 3. Selection preserved by name, not index

The list-message handlers (`changesMsg`, `specsMsg`, `archivedMsg`) currently replace the slice and clamp the index, so a refresh that reorders or removes rows silently moves the cursor. Before replacing each panel's slice, capture the selected item's name; after replacing and sorting, restore `m.sel[panel]` to that name's new index, falling back to the clamped index when the item is gone (e.g. it was just archived). This runs on every list reload, so manual `r` and post-command refreshes benefit too. It composes with `refine-list-panels`' filter-relative indices because it operates where the backing slice is swapped, beneath the filter.

### 4. Focus-regain triggers, guarded

Track `blurred bool` and `lastRefresh time.Time` on the model.

- `tea.BlurMsg` → `blurred = true`. Nothing else.
- `tea.FocusMsg` → refresh only if `blurred` (ignores the focus event some terminals emit right after enabling reporting, and any FocusMsg without a preceding blur), only if `!m.running` (a streaming command's `cmdDoneMsg` already refreshes; interleaving would double-spawn), and only if `time.Since(lastRefresh)` exceeds a short debounce (~1s) so focus flapping while dragging panes around doesn't stack subprocess spawns. Then clear `blurred`, stamp `lastRefresh`, and run `refreshAll`.

Overlays (help, actions, confirm prompt) are unaffected: a refresh only rewrites data state, and the confirm action already captures its target change by name.

### 5. Documentation for terminal support

Focus reporting needs DEC mode 1004: iTerm2, kitty, WezTerm, Alacritty, Ghostty, and Windows Terminal support it; tmux forwards it only with `set -g focus-events on`. Add a README note. Terminals without it emit no events — the feature silently does not engage, and `r` (now a real refresh) remains.

## Risks / Trade-offs

- **[Focus flapping spawns subprocess storms]** → debounce window + `blurred` precondition; a refresh is at most 3 list calls + ~2 preview calls.
- **[Refresh races a running streaming command]** → focus-refresh is skipped while `m.running`; `cmdDoneMsg` refreshes on completion regardless.
- **[Preserved-by-name selection interacts with `refine-list-panels` filtering]** → restoration happens at slice-swap level, beneath the filter's visible-subset mapping; verify with that change's filter tests once both are in.
- **[Carried-over cache entries could mask a deleted artifact]** → the forced `reloadPreview()` re-reads the file/CLI regardless; a deletion arrives as an error/empty message and overwrites the carried entry within one load round-trip.
- **[tmux users get nothing by default]** → README documents `set -g focus-events on`; fallback is today's behavior with a working `r`.
- **[Textual conflicts with in-flight changes]** → sequence this change after `add-live-preview-panes` and `refine-list-panels` are archived.

## Open Questions

- Debounce constant (proposed 1s) — tune by feel once implemented; not spec-relevant.
