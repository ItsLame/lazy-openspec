## Context

lazy-openspec is a Bubble Tea TUI. The root `Model` (`internal/tui/model.go`) tracks a `screen` enum with three values — `screenDashboard`, `screenChangeDetail`, `screenSpecDetail` — plus a `focus panel` that selects one of the three left-column lists (Changes / Specs / Archive). The right pane is a single shared `viewport` (`m.vp`) whose content is rebuilt by `refreshMain()` (`content.go`) based on `screen`.

Current flow:
- On the dashboard, moving the selection re-renders `dashboardPreview` — a short **summary** (`changeSummaryBlock` for changes, a one-line blurb for specs).
- `enterSelected()` switches `screen` to a full-screen detail mode and dispatches loaders (`loadChangeDetail` + `loadArtifact`, or `loadSpecDetail`). `esc` returns to `screenDashboard`.
- Detail keys (`update.go: handleChangeDetailKey`/`handleSpecDetailKey`) forward scroll keys to the viewport and switch artifact tabs / requirements.

Two problems motivate this change (see `proposal.md`):
1. The preview is a summary, not the content, and requires an Enter + screen swap to actually read anything.
2. Selecting an **archived** change hangs on "Loading status…". Archived change names carry a date prefix and live under `openspec/changes/archive/`; `openspec status --change <name>` and `openspec show <name>` do not resolve them (confirmed: `show` returns a validation error because delta specs were synced to main and removed on archive). The `statusMsg`/`changeDetailMsg` handlers only populate their caches on `err == nil` and record nothing on error, so the `content.go` "Loading…" placeholders never clear.

Constraints: single shared viewport; glamour-rendered markdown carries ANSI escapes; archived changes have `proposal.md`/`design.md`/`tasks.md` on disk but **no** `specs/` deltas and are not CLI-resolvable.

## Goals / Non-Goals

**Goals:**
- Render the selected item's full content in the right pane live, as the selection moves, with no Enter required.
- Make Enter a focus toggle between the left list ("nav") and the right preview, so the preview can be scrolled and searched with vim-style keys; `esc` steps back.
- Add an `overview` tab (lifecycle + artifact checklist) as the default preview tab for changes, ahead of `proposal · specs · design · tasks`.
- Fix the archived-item hang by sourcing archived previews from disk and never calling `status`/`show` for them; make active-item load failures resolve to a visible error rather than a permanent "Loading…".
- Add incremental search (`/`, live jump, `n/N`, highlight, `esc`) scoped to the focused preview.

**Non-Goals:**
- No change to the `openspec` CLI or the `internal/openspec` data contracts (only which calls the TUI makes for archived items, plus error propagation into existing message structs).
- No multi-pane/movable selection while the preview is focused — selection changes happen in nav focus only (lazygit-style).
- No regex search, no cross-file search, no fuzzy matching — plain case-insensitive substring only.
- No change to task-toggle semantics beyond the context it fires in (preview focus, tasks tab, non-archived).

## Decisions

### 1. Replace the `screen` enum with a pane-focus field

Drop `screenChangeDetail`/`screenSpecDetail`. Introduce a pane-focus concept:

```go
type pane int
const ( paneNav pane = iota; panePreview )
// Model gains:  activePane pane   // paneNav (default) or panePreview
```

`m.focus panel` is retained for *which* left list is active. The preview's *content* is always derived from the current nav selection (`selectedChange()`/`selectedSpec()`), independent of `activePane`. This is the lazygit model: the left list drives what is previewed; focus only decides who receives keys and the border highlight.

- `renderMain` border highlight becomes `focused := m.activePane == panePreview`.
- Key routing in `handleKey` switches on `m.activePane` instead of `m.screen`: `handleNavKey` (was `handleDashboardKey`) vs `handlePreviewKey` (merges the old change/spec detail handlers).

*Alternative considered:* keep the three screens and just render full content in the dashboard preview. Rejected — the "Enter = focus toggle, not screen swap" requirement and the shared border/hint logic are far simpler with one layout and a focus flag than with three screens that each re-render the same two panes.

### 2. Live-load preview data when the nav selection changes

Selection movement (`moveSel`, panel switch, `1/2/3`, and the initial data-loaded state) calls a new `ensurePreviewLoaded()` that dispatches only the loaders needed for the *current* selection and its *current* tab, if not already cached:

- Selected **spec** → `loadSpecDetail` if `specDetail` for that id isn't loaded.
- Selected **active change** → depends on the active tab: `overview`→`loadStatus`; `proposal/design/tasks`→`loadArtifact` from `openspec/changes/<name>`; `specs`→`loadChangeDetail`.
- Selected **archived change** → `overview`/`proposal/design/tasks`→file reads from `openspec/changes/archive/<name>`; **never** `loadStatus`/`loadChangeDetail`.

The "current change" state (`curChange`, `curChangeDir`, `curArchived`) is recomputed from the selection on every nav move rather than only on Enter. Caches (`statusCache`, `detailCache`, `changeDetail`, `specDetail`) are keyed as today, so revisiting an item is instant.

*Alternative considered:* eagerly load all tabs for the selected change on selection. Rejected — wasteful; load per active tab, lazily, on tab switch (reusing the existing `afterTabChange` cache guard).

### 3. Add `overview` as the first artifact tab

```go
const ( tabOverview artifactTab = iota; tabProposal; tabSpecs; tabDesign; tabTasks; numTabs )
var tabNames = [numTabs]string{"overview", "proposal", "specs", "design", "tasks"}
```

`tabOverview` renders the existing `changeSummaryBlock` (lifecycle + artifact checklist). Default landing tab for a change preview is `tabOverview`. `changeContent` gains a `case tabOverview`. `cacheKey` is unaffected (overview isn't file-backed; it reads `statusCache`).

### 4. Archived previews are file-sourced; load errors resolve the "Loading" state

Two independent fixes so the hang cannot recur:

- **Route:** for `curArchived`, the overview tab is built from on-disk data — task counts parsed from `tasks.md` (via `tasks.Parse`) for the lifecycle line, and artifact presence (which of proposal/design/tasks exist) for the checklist — and the specs tab shows a static note: "Specs were merged into the main specs when this change was archived." No `status`/`show` calls are issued for archived items.
- **Error propagation:** `statusMsg`, `changeDetailMsg`, and `specDetailMsg` handlers record the error (e.g. `m.statusErr[change] = msg.err`, or a sentinel value in the cache) so `changeContent`/`specContent` can render "Failed to load: …" instead of "Loading…". This protects active items too (e.g. a transient CLI error).

A `loadArchivedOverview(changeDir)` command reads `tasks.md` and stats the artifact files, returning a small struct the overview renderer consumes. This keeps file I/O off the Update goroutine.

*Alternative considered:* teach `internal/openspec` to resolve archived names. Rejected — `openspec show` genuinely can't render an archived change (its deltas are gone), so the data simply isn't there; disk is the only source.

### 5. Preview-focus key map

When `activePane == panePreview`:
- Scroll: `j/k`/`↑/↓` (line — but on the **tasks** tab, `j/k` move the task cursor as today), `g`/`G` (top/bottom via `GotoTop`/`GotoBottom`), `ctrl+d`/`ctrl+u` (half-page), `pgup`/`pgdn`.
- Tabs/requirements: `[`/`]` switch artifact tabs for a change; for a spec they jump prev/next requirement (reusing `reqOffsets`, moved off `n/p`).
- `space`: toggle task (tasks tab, non-archived) — unchanged behavior, new trigger context.
- `/`: enter search input (Decision 6). `n`/`N`: next/prev match.
- `esc`: if a search is active, clear it; otherwise return to `paneNav`. `enter`: also returns to `paneNav`.

When `activePane == paneNav`: `j/k` move selection, `tab`/`shift+tab`/`1/2/3` switch panels (as today), `enter` → `panePreview`. Global keys (`q`, `?`, `r`, `v/a/A`, `x`, `ctrl+c`) stay handled before the pane switch, as now.

### 6. Incremental search state machine and highlight

```go
type searchState struct {
    typing  bool     // true while capturing the query line
    query   string
    matches []int    // rendered-content line indices containing a match
    idx     int      // current match within matches
}
// Model gains: search searchState
```

- `/` sets `typing=true`, empties `query`. Rune keys append; `backspace` trims; `enter` confirms (`typing=false`, keep matches); `esc` cancels and clears.
- On every query edit and on content refresh, `recomputeMatches()` scans the **rendered** viewport content: split on `\n`, ANSI-strip each line (via `charmbracelet/x/ansi.Strip`, already in the dependency tree through lipgloss), case-insensitive `strings.Contains`. Record matching line indices; jump the viewport to the first match (`SetYOffset`). `n/N` advance `idx` and `SetYOffset(matches[idx])`.
- **Highlight:** matched lines are re-emitted as ANSI-stripped plain text with each occurrence of the query wrapped in a highlight style (`lipgloss` reverse/inverse; the current `idx` match gets a stronger style). Non-matching lines keep their full glamour styling. Trade-off: a matched line loses its markdown colouring for the duration of the search, in exchange for a reliable inline highlight that works over arbitrary rendered content. The subtitle shows `match i/N` while a search is active.

*Alternative considered:* inject highlight ANSI into the styled line without stripping. Rejected — splicing styles into arbitrary ANSI runs (glamour output) is fragile; strip-and-restyle the matched line is simple and visibly correct, and only affects lines the user is actively looking at.

### 7. View, hints, and help

- `mainTitles()` and `scrollIndicator()`/subtitle account for the two focus states and the search prompt (`/query` + `match i/N`).
- The hint bar (`shortHints`) and help overlay (`keys.go`) gain a preview-focus variant listing scroll/tab/search/`esc` keys, and the nav variant advertises `enter` as "focus preview".

## Risks / Trade-offs

- **Interaction break** (Enter no longer opens a screen; muscle memory) → Documented as BREAKING in the proposal; `esc` still steps back and `enter`/`esc` toggling is discoverable via the updated hint bar and help overlay.
- **Search highlight drops markdown styling on matched lines** → Accepted; only matched lines are affected and only while searching; the rest of the preview keeps its styling. Line-accurate offsets keep `n/N` reliable.
- **Eager loading on every selection move could add latency/spawn commands while scrolling fast** → Loads are cache-guarded and per-active-tab only, and default landing is the cheap `overview` tab; `status`/`show` are skipped entirely for archived items. Bubble Tea batches the `tea.Cmd`s, so rapid movement coalesces.
- **`ansi.Strip` dependency** → It ships transitively with the charmbracelet stack already in `go.mod`; if unavailable, fall back to a small local ANSI-stripping regex.
- **Regression surface** — folding two detail handlers into one preview handler and removing `screen` touches most of the tui package → Covered by targeted tests (Decision-level behaviors: live preview on selection, archived preview loads without hanging, Enter focus toggle, search jump/cycle) plus a manual pass with `go run ./cmd/lazy-openspec`.

## Migration Plan

Pure in-repo TUI refactor; no data migration, no persisted state, no external surface. Ship as one change. Rollback = revert the commit. Because `screen` is removed, verify no lingering references (`grep -rn screenChangeDetail internal/tui`) before building.

## Open Questions

- Should `n/N` remain search-only, or also fall back to requirement navigation for specs when no search is active? Current decision: `[`/`]` owns requirement jumping; `n/N` is search-only, to keep one consistent meaning across previews. Revisit if spec browsing feels worse.
- Should the overview tab also render for a *spec* selection (there is no natural "overview" for a spec)? Current decision: specs have no tab bar; their preview is the requirements list directly.
