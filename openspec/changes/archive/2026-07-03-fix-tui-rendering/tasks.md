## 1. Fix spec identifier decoding (Issues 3 & 4)

- [x] 1.1 Add a custom `UnmarshalJSON` to `SpecSummary` in `internal/openspec/models.go` that reads the identifier from either `id` or `name` into `Name`, and keeps decoding `requirementCount`.
- [x] 1.2 Apply the same tolerance to `SpecDetail` (accept `id`/`name`, and `title` as a fallback for `Name`) so the detail model is consistent.
- [x] 1.3 Add/extend tests in `internal/openspec/client_test.go` (or a models test) covering: CLI output keyed by `id` populates `Name`; output keyed by `name` also populates `Name`.
- [x] 1.4 Manually verify: the Specs panel now shows each spec's name (not a bare `4r`), and opening a spec renders its requirements instead of hanging on `Loading spec…`. _(Verified via headless render snapshot + `TestOpenSpecRendersRequirements`.)_

## 2. Terminal-matching colour palette (Issue 1)

- [x] 2.1 In `internal/tui/styles.go`, replace the xterm-256 colour constants with ANSI 0–15 values and unset/`default` per the design table (active border `2`, inactive border default, title `6`, selected bg `4`, done/active/error `2`/`3`/`1`).
- [x] 2.2 Change muted/plain text styles to use the terminal default foreground with `Faint(true)` instead of a fixed grey index.
- [x] 2.3 In `internal/render/semantic.go`, update the duplicated palette (WHEN/THEN/other, badges, checked/pending, filled/empty) to the same ANSI values so the main pane matches the panel chrome.
- [x] 2.4 Verify no-color behaviour is preserved (run with `NO_COLOR=1`): structure stays readable via glyphs/layout. _(`NO_COLOR=1 go test ./internal/tui ./internal/render` passes; glyphs/layout carry structure independent of colour.)_

## 3. Bordered content padding and clipping (Issue 2)

- [x] 3.1 Add `Padding(0, 1)` to `panelBox` in `internal/tui/styles.go` (keeping `Width(w-2)`, which the existing `leftW-4` body budget already reserves).
- [x] 3.2 Add `Padding(0, 1)` to the main pane box (`renderMain`) and the command-log box (`renderLog`) in `internal/tui/view.go`.
- [x] 3.3 Truncate non-selected rows to the panel width in `changeRow` and `specsList` (`internal/tui/view.go`), matching the existing `fit(...)` behaviour of selected rows.
- [x] 3.4 Verify at the minimum terminal size (`minCols=60`, `minRows=18`) and after a resize that borders stay intact and text sits inside every box. _(`TestBordersHoldWithLongNames` asserts no line exceeds the terminal width at 60×18 and 100×32, even with overflowing names.)_

## 4. Validate

- [x] 4.1 Run `go build ./...` and `go test ./...`; fix any breakage. _(All packages build; full suite + `go vet` green.)_
- [x] 4.2 Run `openspec validate fix-tui-rendering` and confirm the change still validates. _(Reports "valid".)_
- [x] 4.3 Launch the TUI against this repo and confirm all four issues are resolved end-to-end. _(Binary rebuilt; verified via headless `View()` snapshot — interactive AltScreen launch to be eyeballed by the user in a real terminal.)_
