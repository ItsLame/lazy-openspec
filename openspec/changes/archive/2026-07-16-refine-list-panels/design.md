## Context

`add-live-preview-panes` (implemented, validated, **not yet archived**) established the model this change builds on: a `pane` focus (`paneNav` / `panePreview`), a live preview that always follows the nav selection, an artifact tab bar (`overview · proposal · specs · design · tasks`) on `m.tab`, and an incremental `preview-search` over the rendered viewport. All of the keys it defines for moving *within* the preview — `[` / `]`, `/`, `n` / `N` — are gated behind `m.activePane == panePreview`.

The result is that browsing is a three-key round trip (`enter`, `]`, `esc`) and, because `syncSelection` resets `m.tab = tabOverview` whenever the previewed change changes, that round trip has to be repeated for every row. Meanwhile the left column has no filter at all, and the Specs panel appends its `<n>r` count after a variable-length name, so the counts never line up.

The current state of the relevant code:

- `model.go` — `sel [numPanels]int` indexes the **full** `changes` / `specs` / `archived` slices; `selectedChange()` / `selectedSpec()` read straight from those slices; `syncSelection()` resets `tab`, `taskCursor`, search, and scroll when the previewed item changes.
- `update.go` — `handleNavKey` handles panel/selection keys only; `handlePreviewKey` owns `[` / `]`, `/`, `n` / `N`. A guard (`m.activePane == panePreview && m.search.typing`) routes every keystroke into the search query so that global bindings (`q`, `r`, `v`, `x`) are not swallowed while typing.
- `view.go` — `specsList` renders `▪ <name>  <n>r`, truncating the name against a fixed overhead.

## Goals / Non-Goals

**Goals:**

- Make `[` / `]` mean "move between the preview's sections" in **both** panes, so the preview can be steered without ever leaving the list.
- Make the active artifact tab persist as the selection moves, so one tab choice serves a whole browsing pass.
- Make `/` context-sensitive: filter the focused list panel from the list, search the preview from the preview — `lazygit`'s model.
- Give the Specs panel a readable, right-aligned requirement-count gutter on the left.
- Keep selection, preview, and filter mutually coherent: whatever row is highlighted is what the preview shows, filtered or not.

**Non-Goals:**

- Fuzzy or regex matching. Filtering is plain case-insensitive substring matching on the item name, matching the existing preview-search semantics.
- Filtering by anything other than the item name (no lifecycle, percent, or requirement-count predicates).
- Persisting a filter across app restarts, or filtering the preview's content.
- Changing the `internal/openspec` client or the CLI data layer.

## Decisions

### 1. `m.sel` indexes the *visible* items, behind a single accessor

The core structural change. With a filter active, the panel's rows are a subset, and the selection must address that subset — otherwise `j`/`k` would step over hidden rows.

Rather than sprinkling filter checks through `selectedChange`, `selectedSpec`, `moveSel`, `clampSel`, `panelLen`, and the three list renderers, add one accessor:

```go
// visibleChanges/visibleSpecs/visibleArchived apply the active filter (if it
// targets that panel) and return the rows the panel actually renders.
func (m Model) visibleChanges() []openspec.ChangeSummary
func (m Model) visibleSpecs() []openspec.SpecSummary
func (m Model) visibleArchived() []openspec.ChangeSummary
```

Every reader of `m.sel[p]` goes through these; `panelLen(p)` returns the visible length. `m.sel[p]` is then, by definition, an index into the visible slice, and `moveSel`/`clampSel` need no filter awareness at all.

*Alternative considered:* keep `sel` indexing the full slice and map through a `visibleIndices []int` on every read. Rejected — it puts a translation step on every access, which is exactly where an off-by-one hides.

**Risk this creates:** `openspec.ChangeSummary`/`SpecSummary` are rebuilt on every `changesMsg`/`specsMsg`, so filtering is recomputed per render. The lists are tens of items; recomputing a substring match per row per frame is not worth caching.

### 2. The selection follows the *item*, not the index

When the filter query changes (including when it is cleared), the visible slice is rebuilt and a bare index would point at a different item. Before mutating the query, capture the selected item's name; after recomputing the visible set, restore the selection to that name's new index, falling back to `0` when the item is filtered out.

This is what makes typing a query feel like narrowing rather than scrambling: the highlighted row stays highlighted for as long as it matches, and clearing the filter with `esc` leaves you on the row you were on rather than teleporting you.

### 3. One active filter, bound to the panel that opened it

`Model` gains:

```go
type listFilter struct {
    panel  panel  // which panel the filter applies to
    query  string
    typing bool   // query is being edited
}
```

Not a `[numPanels]listFilter` array. A filter left behind on an unfocused panel is invisible state: you would return to Specs, see 2 of 7 rows, and have no on-screen cause. So moving focus to another panel clears the filter.

*Alternative considered:* per-panel persistent filters (closer to `lazygit`, which keeps them and marks the panel title). Rejected for now — it needs a per-panel "filtered" affordance in the border title to be honest, which is more surface than this change warrants. The single-filter shape does not preclude it later.

While a filter is active the panel's border title shows the query (e.g. `2 Specs /nav`), so the narrowed row set always has a visible cause.

### 4. Filter-typing needs its own key-capture guard, before the global bindings

`handleKey` already has this guard for preview search:

```go
if m.activePane == panePreview && m.search.typing { return m.handleSearchInput(msg) }
```

An exactly analogous guard must sit beside it for the list filter:

```go
if m.activePane == paneNav && m.filter.typing { return m.handleFilterInput(msg) }
```

and it **must** precede the common `q` / `?` / `r` / `v` / `a` / `A` / `x` switch. Without it, typing a query containing `q` quits the app and one containing `x` opens the actions overlay. This is the single highest-consequence detail in the change; it gets its own test.

`enter` confirms the query (stops typing, keeps the filter); `esc` clears the filter entirely. A confirmed, non-typing filter still swallows nothing — the global keys work again, and a second `enter` focuses the preview as usual.

### 5. `[` / `]` dispatch on what is previewed, not on which pane is focused

Both `handleNavKey` and `handlePreviewKey` route `[` / `]` to the same two helpers — cycle the artifact tab when a change is previewed, step `reqIdx` through `reqOffsets` when a spec is. The keys therefore have one meaning ("move between the preview's sections") and one implementation, differing only in which pane keeps focus afterwards (the list keeps it; the preview keeps it).

In the **list**, only `[` and `]` are bound — deliberately *not* `left`/`right`. The list already owns a horizontal axis (`h`/`l` cycle panels), so binding the arrows to preview tabs would make `l` and `→` mean different things on the same panel. In the preview, `left`/`right` remain aliases for `[`/`]` as today.

### 6. Sticky tab: drop the reset, keep the rest of `syncSelection`

`syncSelection` stops assigning `m.tab = tabOverview` when the previewed change changes. It still resets `taskCursor`, clears the search, and scrolls to top, because those are *positions within* a document that no longer exists once the document changes; the tab is a *view mode*, which is exactly the thing worth carrying.

`ensurePreviewLoaded` already keys off `m.tab`, so the newly selected change's artifact for the sticky tab loads on arrival with no change to that function.

The tab is not reset when moving between the Changes and Archive panels either — both preview changes and both have the same five tabs, including `overview`. `tabSpecs` on an archived change already renders its static "merged on archive" note, so a sticky `specs` tab degrades gracefully.

### 7. Spec-row gutter width is computed per render from the visible rows

```
 3r  artifact-rendering
12r  change-operations
```

The gutter is `max(len("<n>r"))` over the **visible** specs, right-aligned with `%*s`, rendered faint; then two spaces; then the name, truncated to `width - gutter - 2`. Computing it from the visible set (not the whole list) means a filter that hides the only two-digit spec reclaims that column instead of leaving a ragged blank.

The `▪` bullet is dropped: with a dim count column leading every row, the bullet is a second left-hand anchor competing with the first.

The existing guarantee that a spec row **never** renders as a bare count with an empty name (a real bug this project already fixed once) becomes load-bearing now that the count comes *first* — a truncation bug here would produce exactly that failure mode. The name's truncation budget is floored so the name always gets columns, and the spec keeps that scenario.

## Risks / Trade-offs

- **[Archive order]** This change's deltas modify requirements (`change-navigation`, `spec-navigation`) and a whole capability (`preview-search`) whose current text exists only inside the unarchived `add-live-preview-panes`. If `add-live-preview-panes` is archived *after* this change, its older text overwrites the requirements amended here and the `[`/`]`-from-list and `/`-filter behaviour silently vanishes from the specs. → **Mitigation:** archive `add-live-preview-panes` first; this is stated in the proposal's Impact and is the first task in `tasks.md`.

- **[Selection indexing]** Reindexing `m.sel` against the visible subset touches every selection reader. A missed call site reads the unfiltered slice and previews a row other than the highlighted one. → **Mitigation:** the visible-items accessors are the *only* path to list items — `m.changes`/`m.specs`/`m.archived` are read nowhere else outside them and the loaders. A test asserts that with a filter active, the previewed item matches the highlighted row.

- **[Key capture]** A filter query is free text typed into a key handler whose fallthrough contains `q` (quit) and destructive actions (`A` = archive). A missing or mis-ordered guard is data-destructive, not cosmetic. → **Mitigation:** guard placed before the common-bindings switch, with a test that types a query containing `q`, `r`, `v`, `x`, and `A` and asserts the app neither quits nor runs a command.

- **[Sticky tab surprise]** Landing on a change whose sticky tab is `design` when it has no `design.md` shows an empty state rather than the informative `overview`. → **Trade-off accepted:** this is the cost of the sticky behaviour and the empty state already renders correctly. The `overview` tab remains one `[` away, and the tab bar always shows which tab is active, so the empty pane has a visible cause.

- **[Filter clears on panel switch]** `tab`-ing away from a filtered panel drops the query, so a filter cannot be used to compare two panels side by side. → **Trade-off accepted** in favour of not carrying invisible state (see Decision 3); revisitable by promoting `listFilter` to a per-panel array plus a border-title indicator.
