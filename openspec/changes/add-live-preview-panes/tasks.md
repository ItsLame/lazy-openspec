## 1. Model: pane focus, overview tab, search state

- [x] 1.1 Replace the `screen` enum with a `pane` type (`paneNav`, `panePreview`) and add `activePane pane` to `Model` (default `paneNav`); remove `screenDashboard`/`screenChangeDetail`/`screenSpecDetail` and fix all references.
- [x] 1.2 Add `tabOverview` as the first `artifactTab` and update `tabNames` to `{"overview","proposal","specs","design","tasks"}`; default the previewed change's tab to `tabOverview`.
- [x] 1.3 Add a `searchState` struct (`typing bool`, `query string`, `matches []int`, `idx int`) and a `search searchState` field on `Model`; add helpers to clear it.
- [x] 1.4 Add error-tracking so a failed load resolves the "Loading" state: add `statusErr`/`changeDetailErr`/`specErr` maps (or sentinel cache values) keyed by change/spec.

## 2. Live preview loading on selection

- [x] 2.1 Add `ensurePreviewLoaded()` that, based on the current nav selection and active tab, dispatches only the uncached loader needed (spec → `loadSpecDetail`; active change overview → `loadStatus`; proposal/design/tasks → `loadArtifact`; specs → `loadChangeDetail`).
- [x] 2.2 Recompute `curChange`/`curChangeDir`/`curArchived` from the selection on every nav move (not only on Enter); point `curChangeDir` at `openspec/changes/<name>` or `openspec/changes/archive/<name>`.
- [x] 2.3 Call `ensurePreviewLoaded()` from `handleNavKey` after selection/panel/tab changes and from the data-loaded message handlers (`changesMsg`, `specsMsg`, `archivedMsg`) so the initial preview appears without input.

## 3. Archive: file-sourced preview + no CLI calls

- [x] 3.1 In `ensurePreviewLoaded`/loaders, never issue `loadStatus` or `loadChangeDetail` when `curArchived` is true.
- [x] 3.2 Add a `loadArchivedOverview(changeDir)` command that reads `tasks.md` (task counts via `tasks.Parse`) and stats which artifact files exist, returning a small struct; render the overview tab from it for archived changes.
- [x] 3.3 Render the archived `specs` tab as a static note ("Specs were merged into the main specs when this change was archived.") instead of calling the CLI.
- [x] 3.4 In `statusMsg`/`changeDetailMsg`/`specDetailMsg` handlers, record errors so `changeContent`/`specContent` render "Failed to load: …" instead of a permanent "Loading…".

## 4. Key routing by pane focus

- [x] 4.1 Switch `handleKey` on `m.activePane` instead of `m.screen`; keep global keys (`q`, `?`, `r`, `v/a/A`, `x`, `ctrl+c`) handled first.
- [x] 4.2 Implement `handleNavKey`: `j/k` move selection, `tab`/`shift+tab`/`1`-`3` switch panels, `enter` → `panePreview`; refresh preview + `ensurePreviewLoaded` after moves.
- [x] 4.3 Implement `handlePreviewKey` (merging the old change/spec detail handlers): scroll keys `j/k`, `g/G`, `ctrl+d/u`, `pgup/pgdn`; on the tasks tab `j/k` move the task cursor; `space` toggles a task (non-archived); `[`/`]` switch artifact tabs for a change or jump requirements for a spec; `esc`/`enter` return to `paneNav` (esc first clears an active search).
- [x] 4.4 Ensure switching the previewed item or returning to nav clears `search` and requirement state.

## 5. Overview tab rendering

- [x] 5.1 Add `case tabOverview` to `changeContent` rendering `changeSummaryBlock` for active changes (from `statusCache`) and the file-derived overview for archived changes.
- [x] 5.2 Update `dashboardPreview`/`refreshMain` to always render tabbed change content (default overview) and live spec requirements, driven by the nav selection regardless of `activePane`.

## 6. Incremental search

- [x] 6.1 In `handlePreviewKey`, handle `/` to enter `search.typing`; capture runes, `backspace`, `enter` (confirm), `esc` (cancel/clear).
- [x] 6.2 Add `recomputeMatches()` that splits the rendered viewport content, ANSI-strips each line (`charmbracelet/x/ansi.Strip`, with a regex fallback), case-insensitively finds matching line indices, and jumps to the first match; call it on every query edit and content refresh.
- [x] 6.3 Implement `n`/`N` to advance/rewind `search.idx` with wraparound and `SetYOffset(matches[idx])`.
- [x] 6.4 Highlight matches: re-emit matching lines as ANSI-stripped text with each occurrence styled (current match distinguished); leave non-matching lines' styling intact.

## 7. View, hints, help

- [x] 7.1 Update `renderMain` border highlight to `m.activePane == panePreview`; update `mainTitles`/subtitle to include the overview tab and, when searching, the `/query` prompt and `match i/N` (or "no matches").
- [x] 7.2 Update `shortHints` and the help overlay (`keys.go`) with nav-focus vs preview-focus variants (enter=focus preview; scroll/tab/`/`/`n`,`N`/`esc`).

## 8. Verify

- [x] 8.1 `grep -rn 'screenChangeDetail\|screenSpecDetail\|screenDashboard' internal/tui` returns nothing; `go build ./...` and `go vet ./...` pass.
- [x] 8.2 Add/adjust tests: live preview renders on selection; archived preview loads from disk without hanging; `enter` toggles pane focus and `esc` returns; search jumps to and cycles matches.
- [x] 8.3 Manual pass with `go run ./cmd/lazy-openspec`: select an archived change (no infinite "Loading"), scroll + search a proposal, switch tabs incl. overview, toggle a task on an active change.
- [x] 8.4 `openspec validate add-live-preview-panes --no-interactive` passes.

## 9. Frame-embedded box titles (lazygit-style)

- [x] 9.1 Add a `withBorderTitle(box, title, focused)` helper that rewrites a rendered box's top border to `╭─Title────╮` (title truncated to fit between the corners, frame colour from focus) plus a `borderTitleStyle(focused)` that colours the embedded title with the active border colour when focused and faint otherwise.
- [x] 9.2 Route every box title through the border: drop the title row from `panelBox`, `renderMain`, `renderLog`, and `overlay` bodies, and reclaim the freed line (`bodyH` in `renderLeft`, `vpH` in `layout`, log rows in `renderLog`).
- [x] 9.3 Verify: `go build ./...`, `go vet ./...`, and `go test ./...` pass (incl. `TestBordersHoldWithLongNames`), and a manual render shows titles passing through the borders on all boxes.
