# spec-navigation Specification

## Purpose
TBD - created by archiving change add-lazy-openspec-tui. Update Purpose after archive.
## Requirements
### Requirement: Specs listing
The Specs panel SHALL list the capabilities in the resolved OpenSpec root, each with its identifier and requirement count, sorted for stable ordering. Each row SHALL display the spec's identifier as reported by the CLI; a row SHALL NOT render with a blank name.

#### Scenario: List specs with counts
- **WHEN** the Specs panel renders and the root contains specs
- **THEN** each spec is shown with its name and requirement count (e.g. `auth  4 reqs`)

#### Scenario: Spec identifier is always shown
- **WHEN** the Specs panel renders specs returned by the CLI
- **THEN** each row shows the spec's identifier, and no row renders as only its requirement count (e.g. never a bare `4r` with an empty name)

#### Scenario: Empty specs
- **WHEN** the resolved root has no specs
- **THEN** the Specs panel shows an explicit empty-state message

### Requirement: Requirement and scenario detail
The application SHALL show, for a selected spec, its requirements and each requirement's scenarios in the main pane, with scenarios presented in WHEN/THEN form. Opening a spec SHALL load its detail using the spec's identifier and SHALL reach a rendered or explicit empty state rather than remaining indefinitely on a loading placeholder.

#### Scenario: View a spec's requirements
- **WHEN** the user selects a spec and opens it
- **THEN** the main pane lists each requirement with its description followed by its scenarios rendered as WHEN/THEN pairs

#### Scenario: Opening a spec does not hang
- **WHEN** the user opens a spec from the Specs panel
- **THEN** the application loads that spec's detail by its identifier and renders its requirements, rather than staying on a `Loading spec…` placeholder

### Requirement: Navigate between requirements
The application SHALL let the user move between requirements within a spec detail view (`n` / `p` for next/previous) and scroll long content.

#### Scenario: Jump to next requirement
- **WHEN** the spec detail view is open and the user presses `n`
- **THEN** the view advances to the next requirement, scrolling it into view

