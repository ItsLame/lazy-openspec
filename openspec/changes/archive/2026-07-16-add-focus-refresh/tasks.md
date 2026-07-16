# Tasks: Add Focus Refresh

## 1. Unified refresh path

- [x] 1.1 Add a `reloadPreview()` method to `Model` in `internal/tui/update.go`: a force-load variant of `ensurePreviewLoaded` that dispatches the loaders for the current selection and active tab unconditionally (status/artifact/change-detail/archived-overview for a change; spec detail for a spec), skipping cache-hit checks
- [x] 1.2 Add a `refreshAll()` method that invalidates the client cache, replaces every per-item cache map (`statusCache`, `statusErr`, `detailCache`, `detailErr`, `changeDetail`, `specsErr`, `archivedOv`, `specErr`) with fresh maps while carrying over only the entries backing the current render, and returns a batched `tea.Cmd` of `loadChanges`, `loadSpecs`, `loadArchived`, and `reloadPreview()`
- [x] 1.3 Route the `r` key handler and the `cmdDoneMsg` handler through `refreshAll()` instead of their current invalidate-lists-only logic
- [x] 1.4 Unit test: after `r`, a previously cached artifact's loader is re-dispatched and fresh `artifactMsg` content replaces the cached content (covers "Refresh updates previewed content")

## 2. Selection and scroll preservation

- [x] 2.1 In the `changesMsg`, `specsMsg`, and `archivedMsg` handlers, capture the selected item's name before replacing the slice and restore `m.sel[panel]` to that name's index in the reloaded (sorted) slice, falling back to the existing clamp when the name is gone
- [x] 2.2 Unit test: a reload that reorders the change list keeps the same change selected; a reload where the selected change vanished clamps to a valid row (covers "Refresh preserves selection and scroll" / "Refresh removes a vanished item gracefully")
- [x] 2.3 Verify (test or manual) that the viewport `YOffset` is retained when refreshed content re-renders and clamps when the content shrinks

## 3. Focus events

- [x] 3.1 Add `tea.WithReportFocus()` to the `tea.NewProgram` options in `cmd/lazy-openspec/main.go`
- [x] 3.2 Add `blurred bool` and `lastRefresh time.Time` fields to `Model` in `internal/tui/model.go`
- [x] 3.3 Handle `tea.BlurMsg` in `Update` (set `blurred`; nothing else) and `tea.FocusMsg` (run `refreshAll()` only when `blurred` is set, no streaming command is running, and the ~1s debounce window has elapsed; then clear `blurred` and stamp `lastRefresh`)
- [x] 3.4 Unit tests: blur→focus triggers a refresh; focus without a prior blur does not; blur→focus while `running` does not; two blur→focus cycles inside the debounce window trigger only one refresh (covers the focus-refresh scenarios)

## 4. Documentation and verification

- [x] 4.1 Add a README note on terminal focus-reporting support, including `set -g focus-events on` for tmux
- [x] 4.2 Run `go test ./...` and `openspec validate add-focus-refresh --no-interactive`
- [x] 4.3 Manual check in a tmux split: edit a change's `tasks.md` from the other pane, refocus the TUI pane, confirm the lists and the open preview update without pressing `r` and without the selection or scroll moving
