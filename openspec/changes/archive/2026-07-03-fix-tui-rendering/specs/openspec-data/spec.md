## ADDED Requirements

### Requirement: Tolerant decoding of CLI identifiers
The application SHALL decode the spec identifier from the `openspec` CLI's JSON output tolerantly, accepting the value whether the CLI provides it under the field `id` or the field `name`, so that spec identities populate correctly across CLI versions. A decoded spec SHALL never carry an empty identifier when the CLI supplied one under either field.

#### Scenario: CLI reports specs keyed by id
- **WHEN** `openspec list --specs --json` returns spec entries using the field `id` (as in openspec 1.5.0)
- **THEN** each decoded spec's identifier is populated from that `id`, and it is this identifier that is passed to subsequent commands such as `openspec spec show <id> --json`

#### Scenario: CLI reports specs keyed by name
- **WHEN** a CLI version returns spec entries using the field `name` instead of `id`
- **THEN** the decoded spec's identifier is populated from that `name` without code changes, and the application behaves identically
