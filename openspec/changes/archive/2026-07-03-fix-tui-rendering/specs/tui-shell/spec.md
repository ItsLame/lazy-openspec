## ADDED Requirements

### Requirement: Terminal-matching colour scheme
The application SHALL derive its interface colours from the terminal's ANSI 16-colour palette (colour indices 0–15) and the terminal's `default` (unset) foreground/background, rather than from fixed xterm-256 palette indices, so that the interface matches the user's terminal theme in the manner of `lazygit`/`lazydocker`. The active/focused panel border SHALL use ANSI green, inactive borders SHALL use the terminal default, the selected-line background SHALL use ANSI blue, muted/secondary text SHALL use the terminal default foreground rendered faint, and done/active/error accents SHALL use ANSI green/yellow/red respectively.

#### Scenario: Colours follow a themed terminal
- **WHEN** the application runs in a terminal whose palette has been themed (e.g. Solarized, Gruvbox, or a light profile)
- **THEN** borders, selection highlight, titles, and status glyphs render using that theme's ANSI colours rather than fixed colours that clash with the theme

#### Scenario: Muted text inherits the terminal foreground
- **WHEN** secondary or muted text (hints, empty-state messages, de-emphasised rows) is rendered
- **THEN** it uses the terminal's default foreground colour rendered faint, rather than a hard-coded grey, so it remains legible on both light and dark backgrounds

### Requirement: Bordered content padding and clipping
The application SHALL render the content of every bordered box (the left-column panels, the main pane, and the command-log box) with interior horizontal padding so that text does not touch the box borders, and SHALL truncate list rows to the available inner width so that overlong item names cannot overflow and break the border.

#### Scenario: Content is padded away from borders
- **WHEN** any bordered panel or pane renders its content
- **THEN** there is at least one column of space between the box border and the content on each side, so text sits inside the box rather than flush against the border

#### Scenario: Long item names do not break the border
- **WHEN** a Changes or Specs row whose name is wider than the panel is rendered, whether selected or not
- **THEN** the row is truncated to the panel's inner width (with an ellipsis) and the panel border remains intact and aligned
