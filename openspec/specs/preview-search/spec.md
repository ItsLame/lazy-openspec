# preview-search Specification

## Purpose
TBD - created by archiving change add-live-preview-panes. Update Purpose after archive.
## Requirements
### Requirement: Incremental search within the preview
The application SHALL provide an incremental text search over the focused preview's rendered content, opened with `/`. While the query is being typed the preview SHALL jump to the first match, updating as the query changes. Matching SHALL be plain, case-insensitive substring matching over the visible text (ANSI styling ignored). Pressing `enter` SHALL confirm the query and keep the matches; pressing `esc` SHALL clear the search.

#### Scenario: Open search and jump to first match
- **WHEN** the preview pane is focused and the user presses `/` and types a term
- **THEN** a search prompt shows the query and the preview scrolls to the first line containing the term as it is typed

#### Scenario: Case-insensitive matching
- **WHEN** the query is `preview` and the content contains `Preview`
- **THEN** that line is treated as a match

#### Scenario: Confirm keeps the search
- **WHEN** a search has matches and the user presses `enter`
- **THEN** the query input closes but the matches and highlighting remain, and `n` / `N` continue to cycle them

#### Scenario: Clear search with esc
- **WHEN** a search is active and the user presses `esc`
- **THEN** the search prompt and any match highlighting are removed and focus remains on the preview pane

### Requirement: Cycle and highlight matches
The application SHALL let the user move between matches with `n` (next) and `N` (previous), wrapping around, and SHALL indicate the current match position (e.g. `match i/N`). Matching text SHALL be visually highlighted in the preview, with the current match distinguished from the others.

#### Scenario: Cycle to the next match
- **WHEN** a search has multiple matches and the user presses `n`
- **THEN** the preview scrolls to the next match, the match counter advances, and the current match is highlighted distinctly

#### Scenario: Wrap around at the end
- **WHEN** the current match is the last one and the user presses `n`
- **THEN** the selection wraps to the first match

#### Scenario: No matches
- **WHEN** the query matches no text in the preview
- **THEN** the prompt indicates there are no matches and the preview does not scroll

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

