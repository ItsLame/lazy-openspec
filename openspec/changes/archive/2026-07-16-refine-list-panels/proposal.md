## Why

The two-pane dashboard makes the preview live, but its controls still require a round trip: to read a change's proposal you must press `enter` to focus the preview, press `]` to reach the tab, then `esc` back to keep browsing — and the tab resets to `overview` the moment the selection moves, so skimming the same artifact across several changes means repeating that dance for every row. `/` is likewise stranded in the preview: there is no way to narrow a long list, which is the single most-used key in `lazygit`. Finally, the Specs panel puts its requirement count at the end of a variable-length name, so the counts scatter across the column instead of forming a readable gutter.

## What Changes

- **`[` / `]` switch preview tabs from the list.** The artifact tabs (`overview · proposal · specs · design · tasks`) become steerable while the Changes or Archive panel holds focus — no `enter` first. The keys keep their existing meaning inside the focused preview, so one binding means "move between the preview's sections" in both panes.
- **The active tab persists across selection moves.** **BREAKING** (interaction): moving the selection to another change no longer resets the preview to `overview`. Pick `proposal` once and `j`/`k` down the list to skim every change's proposal. The tab still resets when the previewed *kind* changes (change ↔ spec).
- **`[` / `]` step through requirements from the Specs list.** The Specs panel has no tabs, so the keys keep the meaning they already have for a spec preview — previous/next requirement — scrolling the preview without taking focus off the list.
- **`/` becomes context-sensitive, like `lazygit`.** With the **list** focused, `/` opens an incremental *filter* over the focused panel's rows: non-matching items are hidden as you type, the selection follows the surviving rows, and the preview keeps tracking the selection. With the **preview** focused, `/` remains the incremental *search* over the preview's text. `esc` clears whichever is active.
- **Spec rows lead with a requirement-count gutter.** The Specs panel renders `<n>r` as a dim, right-aligned column on the left, ahead of the name — the way `lazygit` leads a commit row with its timestamp. The `▪` bullet is dropped; the count column becomes the row's left anchor.

## Capabilities

### New Capabilities
- `list-filter`: Incremental filtering of the focused list panel — open with `/`, hide non-matching rows as the query is typed, keep the selection and preview coherent with the filtered set, and clear with `esc`.

### Modified Capabilities
- `change-navigation`: `[` / `]` switch the change preview's artifact tab while the *list* holds focus, not only the preview; the active tab persists as the selection moves between changes instead of resetting to `overview`.
- `spec-navigation`: The Specs panel renders the requirement count as a leading, right-aligned gutter column rather than a trailing suffix; `[` / `]` move between a spec's requirements while the *list* holds focus, not only the preview.
- `preview-search`: `/` is no longer inert while the list holds focus — it opens the list filter there. Search remains scoped to the preview only when the preview holds focus.
- `tui-shell`: The list-focus key map gains `[` / `]` (preview sections) and `/` (filter the focused panel); the hint bar reflects them.

## Impact

- **Depends on `add-live-preview-panes`.** That change is implemented and validated but **not yet archived**, and it is what introduces the tab bar, the pane-focus model, and the `preview-search` capability this change modifies. Its deltas must be archived **before** this one, or archiving `add-live-preview-panes` second would overwrite the requirements amended here.
- **Code:** `internal/tui/` — `model.go` (per-panel filter state; stop resetting `tab` on selection change), `update.go` (route `[`/`]` and `/` in `handleNavKey`; filter-aware selection and `ensurePreviewLoaded`), `search.go` (reuse the matcher for row filtering), `view.go` (spec-row gutter; filter prompt; hint bar), `keys.go` (help text).
- **Selection indices become filter-relative.** `m.sel[panel]` currently indexes the full slice; with a filter active the panels must index the *visible* subset, so every reader of `m.sel` (`selectedChange`, `selectedSpec`, `moveSel`, `clampSel`, the list renderers) has to go through one visible-items accessor. This is the main correctness risk and is detailed in `design.md`.
- **No changes** to the `openspec` CLI data layer or to the `internal/openspec` client contracts.
