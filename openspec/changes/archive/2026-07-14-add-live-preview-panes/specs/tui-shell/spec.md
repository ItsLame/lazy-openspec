## ADDED Requirements

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
The application SHALL render each bordered box's title as part of the box's top border line, in the manner of `lazygit`/`lazydocker` (e.g. `╭─1 Changes────╮`), rather than as a text row inside the box. This SHALL apply to the left-column panels (`1 Changes`, `2 Specs`, `3 Archive`), the main preview pane, the command-log box, and modal overlays. A focused box's embedded title SHALL use the active border colour so the top line reads as one piece; unfocused titles SHALL remain de-emphasised (faint). Titles wider than the frame SHALL be truncated so they can never break the border.

#### Scenario: Panel title passes through the border
- **WHEN** the left column renders the Changes panel
- **THEN** the top border line reads `╭─1 Changes` followed by horizontal border characters, and the first row inside the box is the first list entry rather than the title

#### Scenario: Embedded title follows focus colouring
- **WHEN** the Specs panel holds focus
- **THEN** the `2 Specs` title renders in the active border colour along with its border, while the titles of unfocused boxes render faint

#### Scenario: Overlong title cannot break the frame
- **WHEN** a box title is wider than the box's top border
- **THEN** the title is truncated (with an ellipsis) to fit between the corners and the border remains intact and aligned

## MODIFIED Requirements

### Requirement: Panel focus and navigation
The application SHALL let exactly one panel hold focus at a time, indicate the focused panel with a distinct border/highlight, and allow the user to change focus by cycling with `tab` (and `shift+tab`) or by pressing the panel's number key. Panel cycling and number keys SHALL apply while the list holds focus; pressing `enter` SHALL transfer focus from the list to the preview pane.

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
