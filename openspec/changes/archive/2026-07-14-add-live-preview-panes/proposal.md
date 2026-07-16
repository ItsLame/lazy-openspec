## Why

Today the right pane only shows a short summary until you press Enter, which then swaps the whole view into a separate full-screen "detail" screen. That is an extra step for what should be an at-a-glance browse, and it hides the real content (proposal prose, requirements, tasks) behind a keystroke. Worse, selecting an **archived** change shows "Loading status…" forever: archived names aren't resolvable by `openspec status`/`openspec show`, and the error branches are silent no-ops, so the placeholder never clears. This change makes the second pane a live preview and turns Enter into a focus toggle so the preview can be scrolled and searched with vim-style keys.

## What Changes

- **Live preview on selection.** Moving the cursor over any item in Changes, Specs, or Archive immediately renders its full content in the right pane — no Enter required. The pane updates as the selection moves.
- **Overview tab.** For a change/archived change the preview is tabbed. A new `overview` tab (lifecycle + artifact checklist, today's summary) leads, followed by `proposal · specs · design · tasks`. `overview` is the default landing tab.
- **Enter toggles pane focus, not a screen.** **BREAKING** (interaction): Enter no longer opens a separate detail screen. It moves focus from the left list into the right preview pane; `esc` (or Enter again) returns focus to the list. The two-pane dashboard is now the only layout. The focused pane owns the border highlight and the scroll/search keys.
- **Vim-style scroll + incremental search in the focused preview.** When the preview is focused: `j/k`, `g/G`, `ctrl+d/u`, `pgup/pgdn` scroll; `[`/`]` switch artifact tabs (changes) or jump requirements (specs); `/` opens an incremental search that jumps to matches as you type, `n/N` cycle matches, matches are highlighted, and `esc` clears.
- **Archive preview no longer hangs.** Archived items derive their preview entirely from on-disk artifacts (proposal/design/tasks files, task counts for the overview) and never call `openspec status`/`show`; the specs tab shows a "merged into main on archive" note. Status/detail load failures for active items now resolve to a visible error instead of a permanent "Loading…".

## Capabilities

### New Capabilities
- `preview-search`: Incremental text search within the focused preview pane — open with `/`, live-jump while typing, cycle matches with `n/N`, highlight matches, and clear with `esc`.

### Modified Capabilities
- `change-navigation`: The change preview becomes a live, full-content, tabbed view (new `overview` tab first); Enter toggles focus into the preview pane instead of opening a full-screen detail; archived changes render their preview from disk and no longer hang on "Loading".
- `spec-navigation`: A spec's requirements render live in the preview on selection; Enter focuses the preview pane; next/prev-requirement jumping rebinds to `[`/`]`, freeing `n/N` for search.
- `tui-shell`: Introduce a preview-pane focus alongside the three list panels; the focus indicator and key routing follow the active pane, and the hint bar reflects the two focus states.

## Impact

- **Code:** `internal/tui/` — `model.go` (drop the `screen` enum in favour of a pane-focus field; add `tabOverview`; add search state), `update.go` (route keys by pane focus; eager-load preview data on selection; Enter/esc focus toggle; skip `status`/`show` for archived; resolve load errors), `content.go` (render tabbed live preview incl. overview and archived-from-disk; search highlight), `view.go` (focus indicator, hint bar, search prompt), `commands.go` (archived overview from files; error-carrying messages), `keys.go` (help text per focus).
- **No changes** to the `openspec` CLI data layer contracts beyond how the TUI chooses which source to call for archived items.
- **Behavioral/UX break:** users who relied on Enter opening a separate detail screen and on `esc` returning to a dashboard will find Enter now focuses the preview in place; `esc` still steps back.
