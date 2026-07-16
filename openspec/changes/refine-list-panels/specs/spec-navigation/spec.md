## MODIFIED Requirements

### Requirement: Specs listing
The Specs panel SHALL list the capabilities in the resolved OpenSpec root, each with its identifier and requirement count, sorted for stable ordering. The requirement count SHALL be rendered as a **leading gutter column** — placed before the name, right-aligned to the width of the widest visible count, and de-emphasised — so that counts line up in a readable column down the left edge of the panel rather than trailing names of varying length. Each row SHALL display the spec's identifier as reported by the CLI; a row SHALL NOT render with a blank name.

#### Scenario: List specs with a leading count gutter
- **WHEN** the Specs panel renders and the root contains specs
- **THEN** each row shows its requirement count first as a right-aligned, de-emphasised column (e.g. ` 3r`, `12r`), followed by the spec's name, so the counts align vertically

#### Scenario: Counts of differing width align
- **WHEN** the panel contains both a single-digit count and a double-digit count (e.g. `3r` and `12r`)
- **THEN** the counts are right-aligned to the same column width, so the spec names all begin at the same column

#### Scenario: Spec identifier is always shown
- **WHEN** the Specs panel renders specs returned by the CLI
- **THEN** each row shows the spec's identifier, and no row renders as only its requirement count (e.g. never a bare `4r` with an empty name)

#### Scenario: Long names do not displace the gutter
- **WHEN** a spec's name is wider than the space remaining after the count gutter
- **THEN** the name is truncated with an ellipsis, the count gutter remains intact and aligned, and the panel border is not broken

#### Scenario: Empty specs
- **WHEN** the resolved root has no specs
- **THEN** the Specs panel shows an explicit empty-state message

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
The application SHALL let the user move between requirements within the previewed spec using `[` / `]` for previous/next, from **either** pane: while the Specs panel holds focus (scrolling the preview without taking focus off the list) and while the preview pane holds focus. Long content SHALL be scrollable with vim-style keys in the focused preview. The `n` / `N` keys SHALL be reserved for search-match navigation rather than requirement navigation.

#### Scenario: Jump to next requirement from the preview
- **WHEN** the preview pane is focused for a spec and the user presses `]`
- **THEN** the view advances to the next requirement, scrolling it into view

#### Scenario: Jump to next requirement from the list
- **WHEN** a spec is selected, the Specs panel holds focus, and the user presses `]`
- **THEN** the preview scrolls to the next requirement while keyboard focus remains on the list, so `j` / `k` still move the spec selection

#### Scenario: Requirement keys do not shadow search
- **WHEN** the preview pane is focused for a spec
- **THEN** `[` / `]` move between requirements and `n` / `N` are handled by search, not requirement navigation
