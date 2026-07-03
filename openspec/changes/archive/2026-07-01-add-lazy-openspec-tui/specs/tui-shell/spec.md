## ADDED Requirements

### Requirement: Multi-panel layout
The application SHALL present a `lazygit`-style layout consisting of a left column of stacked, labelled panels (Changes, Specs, Archive), a main content pane on the right, and a bottom bar that combines a command log with a keybinding hint line.

#### Scenario: Initial render
- **WHEN** the user launches `lazy-openspec` in a directory containing an `openspec/` root
- **THEN** the left column renders the Changes, Specs, and Archive panels, the main pane shows the selection preview, and the bottom bar shows context-appropriate keybindings

#### Scenario: Terminal resize
- **WHEN** the terminal is resized while the app is running
- **THEN** the panels reflow to the new dimensions without overlap or truncation of borders

### Requirement: Panel focus and navigation
The application SHALL let exactly one panel hold focus at a time, indicate the focused panel with a distinct border/highlight, and allow the user to change focus by cycling with `tab` (and `shift+tab`) or by pressing the panel's number key.

#### Scenario: Cycle focus with tab
- **WHEN** the Changes panel is focused and the user presses `tab`
- **THEN** focus moves to the next panel and its border highlight updates accordingly

#### Scenario: Jump to panel by number
- **WHEN** the user presses `2`
- **THEN** the Specs panel receives focus regardless of which panel was previously focused

#### Scenario: Move selection within a panel
- **WHEN** a panel is focused and the user presses the down arrow or `j`
- **THEN** the selection highlight moves to the next item in that panel and the main pane preview updates to the newly selected item

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
