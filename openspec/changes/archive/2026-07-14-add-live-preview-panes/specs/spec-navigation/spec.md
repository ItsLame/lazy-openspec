## MODIFIED Requirements

### Requirement: Requirement and scenario detail
The application SHALL show, for the selected spec, its requirements and each requirement's scenarios in the main pane, with scenarios presented in WHEN/THEN form, rendered live as the selection moves without requiring `enter`. Selecting a spec SHALL load its detail using the spec's identifier and SHALL reach a rendered or explicit empty state rather than remaining indefinitely on a loading placeholder. Pressing `enter` SHALL move keyboard focus into the preview pane so the requirements can be scrolled and searched.

#### Scenario: View a spec's requirements
- **WHEN** a spec is selected in the Specs panel
- **THEN** the main pane immediately lists each requirement with its description followed by its scenarios rendered as WHEN/THEN pairs, with no `enter` required

#### Scenario: Selecting a spec does not hang
- **WHEN** a spec is selected in the Specs panel
- **THEN** the application loads that spec's detail by its identifier and renders its requirements, rather than staying on a `Loading spec…` placeholder

#### Scenario: Focus the spec preview with enter
- **WHEN** the user presses `enter` on a selected spec
- **THEN** keyboard focus moves to the preview pane and scroll/search keys are routed to it

### Requirement: Navigate between requirements
The application SHALL let the user move between requirements within the focused spec preview using `[` / `]` for previous/next, and scroll long content with vim-style keys. The `n` / `N` keys SHALL be reserved for search-match navigation rather than requirement navigation.

#### Scenario: Jump to next requirement
- **WHEN** the preview pane is focused for a spec and the user presses `]`
- **THEN** the view advances to the next requirement, scrolling it into view

#### Scenario: Requirement keys do not shadow search
- **WHEN** the preview pane is focused for a spec
- **THEN** `[` / `]` move between requirements and `n` / `N` are handled by search, not requirement navigation
