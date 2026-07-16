## 1. Prerequisite: land the change this stacks on

- [x] 1.1 Land `add-live-preview-panes` (finish its open task 9.3, commit the `internal/tui` work) and archive it with `openspec archive add-live-preview-panes -y`, so its `preview-search` capability and its `change-navigation` / `spec-navigation` / `tui-shell` deltas are in the main specs **before** this change's deltas modify them.
- [x] 1.2 Confirm `openspec/specs/preview-search/spec.md` exists and that `change-navigation`'s "Drill into a change and switch artifacts" reflects the pane-focus wording; then re-run `openspec validate refine-list-panels --no-interactive`.

## 2. Sticky artifact tab

- [x] 2.1 In `syncSelection` (`model.go`), stop assigning `m.tab = tabOverview` when the previewed change changes; keep the `taskCursor`, search-clear, and `vp.GotoTop()` resets.
- [x] 2.2 Verify `ensurePreviewLoaded` (`update.go`) already dispatches for the sticky `m.tab` on the newly selected change (it keys off `m.tab`, so no change expected) and that a sticky tab pointing at a missing artifact renders the existing empty state rather than a permanent "Loading…".
- [x] 2.3 Test: with `tab == tabProposal`, moving the selection to another change previews *that* change's proposal and leaves `tab == tabProposal`.

## 3. `[` / `]` from the list

- [x] 3.1 Extract the tab-cycling and requirement-jumping bodies out of `previewChangeKey` / `previewSpecKey` into two pane-agnostic helpers (e.g. `cycleTab(delta)` and `jumpRequirement(delta)`) so both key handlers call one implementation.
- [x] 3.2 In `handleNavKey`, add `[` and `]`: when a change is selected call `cycleTab` then `afterTabChange` (which reloads and refreshes); when a spec is selected call `jumpRequirement`. Leave focus on `paneNav` in both cases.
- [x] 3.3 Do **not** bind `left` / `right` in `handleNavKey` — they must stay unbound there so they cannot collide with `h` / `l` panel cycling. `previewChangeKey` / `previewSpecKey` keep their `left` / `right` aliases.
- [x] 3.4 Test: `]` on a list-focused change advances the tab and leaves `activePane == paneNav`; `]` on a list-focused spec advances `reqIdx` and moves the viewport offset.

## 4. List filter: state and key capture

- [x] 4.1 Add a `listFilter` struct (`panel panel`, `query string`, `typing bool`) and a `filter listFilter` field to `Model`, plus `active()` / `clearFilter()` helpers mirroring `searchState`.
- [x] 4.2 In `handleKey`, add the guard `if m.activePane == paneNav && m.filter.typing { return m.handleFilterInput(msg) }` **immediately beside the existing preview-search guard and before the common `q`/`?`/`r`/`v`/`a`/`A`/`x` switch**.
- [x] 4.3 Implement `handleFilterInput`: runes and `space` append to the query, `backspace` deletes, `enter` confirms (`typing = false`, filter kept), `esc` clears the filter entirely. Re-select and refresh after every edit (task 5.3).
- [x] 4.4 In `handleNavKey`, bind `/` to open the filter on `m.focus` (`m.filter = listFilter{panel: m.focus, typing: true}`) and `esc` to clear an active filter.
- [x] 4.5 Clear the filter whenever list focus moves to another panel (the `tab` / `shift+tab` / `h` / `l` / `1`-`3` cases in `handleNavKey`).
- [x] 4.6 Test (highest value): with a filter being typed, feeding the keys `q`, `r`, `v`, `x`, `A` appends them to the query and does **not** quit, reload, validate, archive, or open the actions overlay; after `enter`, `q` quits again.

## 5. List filter: visible-items model

- [x] 5.1 Add `visibleChanges()`, `visibleSpecs()`, `visibleArchived()` on `Model`, each applying `m.filter` only when `m.filter.panel` matches that panel, using case-insensitive substring matching on the item name.
- [x] 5.2 Reroute every reader of the raw slices through them: `selectedChange`, `selectedSpec`, `panelLen`, `changesList`, `specsList`, `archiveList`. After this, `m.sel[p]` indexes the **visible** slice and `moveSel` / `clampSel` need no filter awareness.
- [x] 5.3 Add `reselectAfterFilter(prevName string)`: before a query edit capture the selected item's name; after recomputing the visible set restore the selection to that name's new index, falling back to `0` when it no longer matches. Call `syncSelection` + `ensurePreviewLoaded` afterwards so the preview follows.
- [x] 5.4 Handle the empty-result case: `panelLen == 0` means nothing is selected, so `selectedChange`/`selectedSpec` return `false` and the preview renders its "nothing selected" empty state instead of stale content.
- [x] 5.5 In `changesList`, only emit a lifecycle group header when that group still has at least one visible row, and compute `focusLine` from the visible index.
- [x] 5.6 Test: with a filter active the previewed item matches the highlighted row; `j` steps between matching rows only; clearing with `esc` restores the full list with the same item still selected.

## 6. Spec-row count gutter

- [x] 6.1 Rewrite `specsList` (`view.go`): compute the gutter width as the widest `"<n>r"` over the **visible** specs, render it right-aligned (`%*s`) and faint as the leading column, then two spaces, then the name; drop the `▪` bullet.
- [x] 6.2 Truncate the name to `width - gutter - 2` with a floor so the name always gets columns; keep the selected row rendered through `selectedItem.Render(fit(...))` across the full width.
- [x] 6.3 Test: counts of differing width right-align to a common column and names all start at the same column; a name longer than the panel is ellipsised without breaking the border; no row ever renders as a bare count with an empty name.

## 7. View, hints, help

- [x] 7.1 Show the active filter query on the filtered panel (border title, e.g. `2 Specs /nav`, with the typing caret while `filter.typing`).
- [x] 7.2 Render an explicit no-matches message in a panel whose filter matches nothing, instead of an empty box.
- [x] 7.3 Update `shortHints` for list focus to advertise `[ ]` (preview sections) and `/` (filter), and update the help overlay in `keys.go` with the list-focus vs preview-focus key sets.

## 8. Verify

- [x] 8.1 `go build ./...`, `go vet ./...`, and `go test ./...` pass, including the existing `TestBordersHoldWithLongNames`.
- [x] 8.2 Manual pass with `go run ./cmd/lazy-openspec`: from the list, `]` walks the tabs; pick `proposal` and `j`/`k` down the Changes list skimming each proposal; `/` in Specs narrows the list live and `esc` restores it with the selection intact; the Specs gutter aligns; `/` in a focused preview still searches the preview.
- [x] 8.3 `openspec validate refine-list-panels --no-interactive` passes.
