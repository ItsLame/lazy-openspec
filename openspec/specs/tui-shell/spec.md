# tui-shell Specification

## Purpose
TBD - created by archiving change add-lazy-openspec-tui. Update Purpose after archive.
## Requirements
### Requirement: Multi-panel layout
The application SHALL present a `lazygit`-style layout consisting of a left column of stacked, labelled panels (Changes, Specs, Archive), a main content pane on the right, and a bottom bar that combines a command log with a keybinding hint line.

#### Scenario: Initial render
- **WHEN** the user launches `lazy-openspec` in a directory containing an `openspec/` root
- **THEN** the left column renders the Changes, Specs, and Archive panels, the main pane shows the selection preview, and the bottom bar shows context-appropriate keybindings

#### Scenario: Terminal resize
- **WHEN** the terminal is resized while the app is running
- **THEN** the panels reflow to the new dimensions without overlap or truncation of borders

### Requirement: Panel focus and navigation
The application SHALL let exactly one panel hold focus at a time, indicate the focused panel with a distinct border/highlight, and allow the user to change focus by cycling with `tab` (and `shift+tab`) or by pressing the panel's number key. Panel cycling and number keys SHALL apply while the list holds focus; pressing `enter` SHALL transfer focus from the list to the preview pane. While the list holds focus, the application SHALL additionally bind `[` / `]` to move between the preview's sections (the artifact tabs of a previewed change, or the requirements of a previewed spec) without transferring focus, and `/` to filter the focused panel's rows. The horizontal arrow keys SHALL NOT be bound to preview sections while the list holds focus, so that they do not conflict with `h` / `l` panel cycling.

#### Scenario: Cycle focus with tab
- **WHEN** the Changes panel is focused and the user presses `tab`
- **THEN** focus moves to the next panel and its border highlight updates accordingly

#### Scenario: Jump to panel by number
- **WHEN** the user presses `2` while the list holds focus
- **THEN** the Specs panel receives focus regardless of which panel was previously focused

#### Scenario: Move selection within a panel
- **WHEN** a panel is focused and the user presses the down arrow or `j`
- **THEN** the selection highlight moves to the next item in that panel and the main pane preview updates to the newly selected item

#### Scenario: Enter transfers focus to the preview
- **WHEN** a list panel holds focus and the user presses `enter`
- **THEN** keyboard focus moves to the preview pane and the panel keys (`tab`, `1`–`3`) no longer move list focus until focus returns

#### Scenario: Preview sections are steerable from the list
- **WHEN** a list panel holds focus and the user presses `[` or `]`
- **THEN** the preview moves to the previous/next section for the selected item, and keyboard focus remains on the list

#### Scenario: Hint bar reflects the list-focus keys
- **WHEN** the list holds focus
- **THEN** the hint bar advertises the list-context keys, including `[` `]` for the preview's sections and `/` to filter the panel

### Requirement: Help overlay
The application SHALL provide a help overlay, opened with `?`, that lists the available keybindings for the current context and is dismissible with `?`, `esc`, or `q`.

#### Scenario: Open and close help
- **WHEN** the user presses `?`
- **THEN** an overlay listing the current keybindings is displayed, and pressing `esc` closes it and returns focus to the previously focused panel

### Requirement: Quit
The application SHALL exit cleanly, restoring the terminal state, when the user presses `q` or `ctrl+c` outside of any modal input.

#### Scenario: Quit from the dashboard
- **WHEN** the user presses `q` while a panel is focused and no overlay or input is active
- **THEN** the application exits and the terminal is restored to its normal (non-alternate-screen) state

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

### Requirement: Preview pane focus and scrolling
The application SHALL support moving keyboard focus between the list and the preview pane. When the preview pane holds focus, its border SHALL be highlighted as active, the hint bar SHALL show preview-context keys, and vim-style keys SHALL scroll the preview: `j`/`k` (line, except the tasks tab where they move the task cursor), `g`/`G` (top/bottom), `ctrl+d`/`ctrl+u` (half-page), and `pgup`/`pgdn`. Pressing `esc` with no active search SHALL return focus to the list.

#### Scenario: Scroll the focused preview
- **WHEN** the preview pane is focused and the user presses `ctrl+d`
- **THEN** the preview scrolls down by half a page and the scroll indicator updates

#### Scenario: Focus indicator follows the active pane
- **WHEN** focus is on the preview pane
- **THEN** the preview pane's border is highlighted as active and the list panel borders are not

#### Scenario: Esc returns focus to the list
- **WHEN** the preview pane is focused, no search is active, and the user presses `esc`
- **THEN** focus returns to the list with the same item still selected

### Requirement: Frame-embedded box titles
The application SHALL render each bordered box's title as part of the box's top border line, in the manner of `lazygit`/`lazydocker` (e.g. `╭─[1]─Changes────╮`), rather than as a text row inside the box. This SHALL apply to the left-column panels (`[1]─Changes`, `[2]─Specs`, `[3]─Archive`), the main preview pane, the command-log box, and modal overlays. Each left-column panel's title SHALL carry its jump key in square brackets, joined to the panel name by the frame's horizontal border character (`─`, U+2500) rather than a hyphen, matching how `lazygit` draws its jump labels. A focused box's embedded title SHALL use the active border colour so the top line reads as one piece; unfocused titles SHALL remain de-emphasised (faint). Titles wider than the frame SHALL be truncated so they can never break the border.

#### Scenario: Panel title passes through the border
- **WHEN** the left column renders the Changes panel
- **THEN** the top border line reads `╭─[1]─Changes` followed by horizontal border characters, and the first row inside the box is the first list entry rather than the title

#### Scenario: Embedded title follows focus colouring
- **WHEN** the Specs panel holds focus
- **THEN** the `[2]─Specs` title renders in the active border colour along with its border, while the titles of unfocused boxes render faint

#### Scenario: Overlong title cannot break the frame
- **WHEN** a box title is wider than the box's top border
- **THEN** the title is truncated (with an ellipsis) to fit between the corners and the border remains intact and aligned

