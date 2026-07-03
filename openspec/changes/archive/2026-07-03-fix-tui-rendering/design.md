## Context

`lazy-openspec` is a Bubble Tea + lipgloss TUI that reads data by shelling out to the `openspec` CLI (`--json`). Four rendering defects were diagnosed:

1. **Colours don't match the terminal.** The palette in `internal/tui/styles.go` (and a duplicated palette in `internal/render/semantic.go`) uses fixed xterm-256 cube indices (`"240"`, `"39"`, `"246"`, `"214"`, `"42"`, `"203"`, `"236"`, `"255"`). These render identically regardless of the user's terminal theme, so they clash on themed/light terminals.
2. **Text hugs the borders.** The three bordered boxes — `panelBox` (`styles.go`), the main pane (`view.go: renderMain`), and the command log (`view.go: renderLog`) — set a border but no padding.
3. **Specs render as `4r` with no name**, and **4. opening a spec hangs on `Loading spec…`** — a single root cause. `SpecSummary` decodes JSON field `name` (`internal/openspec/models.go`), but the CLI emits `id`. `Name` is therefore always empty, so the row shows only the requirement count, and the empty name is passed as the spec id to `openspec spec show "" --json`, which errors; `specDetailMsg` only sets `specDetail` on success, so the view stays on its loading placeholder forever.

Constraints: presentation/decoding only — no change to keybindings, commands, CLI invocations, or the openspec contract. lipgloss is v1.1.x (ANSI colours, `AdaptiveColor`, and unset/`default` colours all available).

## Goals / Non-Goals

**Goals:**
- Colours derive from the terminal's ANSI 16-colour palette + `default`, mirroring `lazygit`/`lazydocker`, so the UI follows the user's theme.
- Panel/pane/log content sits inside its box with interior padding.
- Spec rows always show their identifier; opening a spec loads its requirements instead of hanging.
- Keep the two colour palettes (chrome and semantic renderer) in visual agreement.

**Non-Goals:**
- No user-configurable theme file (unlike lazygit's `gui.theme`); a single terminal-matching palette is enough for now.
- No embedding of panel titles into the top border (lazygit's `╭─ Changes ─╮`); optional polish, out of scope.
- No change to Glamour prose styling (its auto-style already adapts light/dark).
- No migration off the deprecated `openspec spec show` command (`trimToJSON` already tolerates its warning line).

## Decisions

### 1. Terminal-matching palette via ANSI 0–15 + `default`
Replace the xterm-256 indices with ANSI 16-colour indices and unset/`default`, mirroring lazygit's default theme:

| Role | Now (256) | New |
| --- | --- | --- |
| active/focus border | `39` | `2` (green) |
| inactive border | `240` | unset / `default` |
| title / accent | `39` | `6` (cyan) |
| muted / plain text | `246` | unset + `Faint(true)` |
| selected line bg | `236` (fg `255`) | `4` (blue) bg, `default`/`15` fg |
| done / active / error | `42` / `214` / `203` | `2` / `3` / `1` |

Applied in **both** `styles.go` and `semantic.go` (semantic keeps its own copies of `42`/`39`/`214`/`203`); both must change together or the main pane and panel chrome will disagree.

- **Why ANSI 0–15 over 256-cube:** indices 0–15 are remapped by the terminal to the user's theme; the 256-cube is absolute. This is exactly how lazygit "matches the terminal."
- **Why `default`/faint for muted text over a grey index:** a fixed grey is unreadable on some themes; the terminal default foreground rendered faint always has correct contrast.
- **Alternative considered — `AdaptiveColor{Light, Dark}`:** still hard-codes specific colours per mode; it adapts to background lightness but not to the user's actual palette. Rejected in favour of true ANSI colours.

### 2. Add `Padding(0, 1)` to the three bordered boxes
The existing width budgets already reserve the space: `renderLeft` builds bodies at `leftW-4` while the box sets `Width(leftW-2)`; `renderMain` builds the viewport at `mainW-4` while the box sets `Width(mainW-2)`. Those `-4`s only make sense with 1 column of padding per side (`-2` border, `-2` padding). Adding `Padding(0, 1)` consumes exactly that reserved budget, so text moves inside with no other math change.

- **Alternative considered — widen content instead of padding:** would make text touch the border, the opposite of the goal.

### 3. Truncate non-selected rows, not just selected ones
Today only selected rows go through `fit(...)`; non-selected `changeRow` and `specsList` rows are emitted untruncated. Empty spec names masked this, but real names (e.g. `artifact-rendering`, 18 chars) in a 26-wide panel can overflow and break the border box. Apply the same width-clamp to both branches.

### 4. Tolerant identifier decoding (`id` or `name`)
Give `SpecSummary` (and `SpecDetail`) a custom `UnmarshalJSON` that reads the identifier from whichever of `id` / `name` the CLI supplies, into the existing `Name` field. Every call site already uses `.Name`, so no view code changes; the populated name flows correctly as the spec id into `spec show`.

- **Why tolerant over a one-line retag to `json:"id"`:** the `SpecDetail` comment already notes CLI versions differ on `id`/`name`; tolerant decoding survives that drift. Cost is a small `UnmarshalJSON` and a test.

## Risks / Trade-offs

- **Some ANSI themes make green/blue low-contrast** → keep structural glyphs (`◉`, `✓`, `☐`) so meaning survives without colour; the no-color requirement already mandates monochrome legibility.
- **`Padding` interacts with `MaxWidth`/`MaxHeight` clipping** → verify at the minimum terminal size (`minCols=60`, `minRows=18`) after the change; the reserved budget should keep totals unchanged, but a resize check confirms borders don't truncate.
- **Truncation could hide the tail of long names** → acceptable and expected (ellipsis); selected rows already behaved this way.
- **Tolerant decode masking genuinely absent identifiers** → the requirement only promises a non-empty id when the CLI supplied one; if both fields are absent the id is legitimately empty and upstream error handling applies.

## Open Questions

- None blocking. Optional future polish (title-in-border, configurable theme) is deliberately deferred.
