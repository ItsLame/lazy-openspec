## MODIFIED Requirements

### Requirement: Search is scoped to the focused preview
Preview search SHALL be available only while the preview pane holds focus, and SHALL operate only over the current preview content. Changing the previewed item or returning focus to the list SHALL clear any active search. While the **list** holds focus, `/` SHALL NOT open a preview search — it opens the list filter over the focused panel's rows instead (see the `list-filter` capability). The two are mutually exclusive: at most one of a preview search and a list filter is active at a time.

#### Scenario: Slash from the list filters instead of searching
- **WHEN** the list holds focus and the user presses `/`
- **THEN** the list filter opens over the focused panel's rows, and no preview search prompt opens

#### Scenario: Slash from the preview searches
- **WHEN** the preview pane holds focus and the user presses `/`
- **THEN** the preview search prompt opens over the preview's content, and the list is not filtered

#### Scenario: Search clears when leaving the preview
- **WHEN** a search is active and focus returns to the list
- **THEN** the search state is cleared so it does not persist onto the next previewed item
