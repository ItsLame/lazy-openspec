## Why

The TUI has four rendering defects that make it look and behave unlike its `lazygit`/`lazydocker` inspiration: hard-coded 256-colour indices ignore the user's terminal theme, panel text hugs the borders, every spec renders as a bare requirement count (e.g. `4r`) with no name, and opening a spec hangs on `Loading spec…` forever. The last two share a single root cause — the spec decoder reads the JSON field `name`, but the `openspec` CLI emits `id` — so a blank name is passed back as the spec id, and `openspec spec show "" --json` errors out.

## What Changes

- **Terminal-matching colours**: Replace the fixed xterm-256 palette (in both `internal/tui/styles.go` and `internal/render/semantic.go`) with the terminal's ANSI 16-colour palette (0–15) plus unset/`default`, mirroring `lazygit` — green active border, `default` inactive border, blue selected line, `default`+faint muted text, and `2`/`3`/`1` for done/active/error. Colours now follow the user's theme instead of overriding it.
- **Text inside the box**: Add interior padding to the three bordered boxes (left panels, main pane, command log) so content no longer touches the border. The existing width budgets already reserve this space.
- **Prevent border breakage**: Truncate non-selected list rows to the panel width (only selected rows are truncated today), so real spec/change names cannot overflow and break the box.
- **Correct spec identity decoding**: Decode the CLI's spec identifier tolerantly, accepting either `id` or `name`. This fixes both the `4r`-with-no-name display **and** the stuck `Loading spec…` (the empty name was being passed as the spec id).

No behavioural or CLI-contract changes — this is a presentation and JSON-decoding fix only.

## Capabilities

### New Capabilities

_None — all changes modify existing capabilities._

### Modified Capabilities

- `tui-shell`: Panel/box chrome SHALL derive colours from the terminal's ANSI palette and SHALL pad content away from borders and truncate list rows so borders never break.
- `artifact-rendering`: The semantic renderer's colours SHALL come from the terminal ANSI palette so the main pane matches the panel chrome and the user's theme.
- `openspec-data`: Decoding of CLI JSON SHALL tolerate the spec identifier appearing as either `id` or `name`, so spec names populate correctly across CLI versions.
- `spec-navigation`: The Specs panel SHALL show each spec's identifier (never blank), and opening a spec SHALL load its requirement detail rather than hanging.

## Impact

- **Code**: `internal/tui/styles.go`, `internal/tui/view.go`, `internal/render/semantic.go`, `internal/openspec/models.go` (and their tests).
- **Behaviour**: Purely visual/decoding; no keybindings, commands, or CLI invocations change.
- **Dependencies**: None added — uses existing `lipgloss` ANSI colour support.
