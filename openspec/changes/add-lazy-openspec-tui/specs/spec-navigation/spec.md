## ADDED Requirements

### Requirement: Specs listing
The Specs panel SHALL list the capabilities in the resolved OpenSpec root, each with its requirement count, sorted for stable ordering.

#### Scenario: List specs with counts
- **WHEN** the Specs panel renders and the root contains specs
- **THEN** each spec is shown with its name and requirement count (e.g. `auth  4 reqs`)

#### Scenario: Empty specs
- **WHEN** the resolved root has no specs
- **THEN** the Specs panel shows an explicit empty-state message

### Requirement: Requirement and scenario detail
The application SHALL show, for a selected spec, its requirements and each requirement's scenarios in the main pane, with scenarios presented in WHEN/THEN form.

#### Scenario: View a spec's requirements
- **WHEN** the user selects a spec and opens it
- **THEN** the main pane lists each requirement with its description followed by its scenarios rendered as WHEN/THEN pairs

### Requirement: Navigate between requirements
The application SHALL let the user move between requirements within a spec detail view (`n` / `p` for next/previous) and scroll long content.

#### Scenario: Jump to next requirement
- **WHEN** the spec detail view is open and the user presses `n`
- **THEN** the view advances to the next requirement, scrolling it into view
