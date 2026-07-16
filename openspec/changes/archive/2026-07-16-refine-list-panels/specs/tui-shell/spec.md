## MODIFIED Requirements

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
