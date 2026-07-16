## ADDED Requirements

### Requirement: Filter the focused list panel
The application SHALL provide an incremental filter over the focused list panel, opened with `/` while the list holds focus. As the query is typed, the panel SHALL show only the items whose name matches, hiding the rest. Matching SHALL be plain, case-insensitive substring matching against the item's name. Pressing `enter` SHALL confirm the query and keep the panel filtered; pressing `esc` SHALL clear the filter and restore the full list.

#### Scenario: Open the filter and narrow the list
- **WHEN** the Specs panel holds focus and the user presses `/` and types `nav`
- **THEN** a filter prompt shows the query and the Specs panel shows only specs whose name contains `nav`, hiding the others as the query is typed

#### Scenario: Case-insensitive matching
- **WHEN** the query is `NAV` and the panel contains `spec-navigation`
- **THEN** that row is shown as a match

#### Scenario: Confirm keeps the filter
- **WHEN** a filter has matches and the user presses `enter`
- **THEN** the query input closes, the panel stays narrowed to the matching rows, and the list keys (`j` / `k`, `[` / `]`) operate on the filtered rows

#### Scenario: Clear the filter with esc
- **WHEN** a filter is active and the user presses `esc`
- **THEN** the filter is cleared, the panel shows all of its items again, and focus remains on the list

#### Scenario: No matches
- **WHEN** the query matches no item in the focused panel
- **THEN** the panel shows an explicit no-matches message rather than an empty box, and the preview pane shows an empty state rather than stale content

### Requirement: Filter query captures keystrokes
While a filter query is being typed, the application SHALL route every keystroke into the query rather than to the global bindings, so that letters bound to global actions (such as `q` to quit, `r` to reload, `v` to validate, `x` for the actions overlay, or `A` to archive) are inserted as text instead of triggering their action.

#### Scenario: Global keys are typed, not triggered
- **WHEN** a filter query is being typed and the user types characters including `q`, `r`, `v`, `x`, and `A`
- **THEN** those characters are appended to the query, and the application does not quit, reload, validate, archive, or open the actions overlay

#### Scenario: Global keys work again once confirmed
- **WHEN** the user presses `enter` to confirm a filter and then presses `q`
- **THEN** the application quits, because the query is no longer capturing keystrokes

### Requirement: Selection and preview stay coherent with the filter
The selection SHALL address only the visible (matching) rows: moving the selection SHALL step between matching rows without landing on hidden ones, and the preview SHALL always show the currently highlighted row. When the query changes, the selection SHALL follow the previously selected item to its new position if that item still matches, and otherwise SHALL fall back to the first visible row.

#### Scenario: Selection steps over hidden rows
- **WHEN** a filter is active and the user presses `j`
- **THEN** the selection moves to the next *matching* row, and the preview updates to that row's content

#### Scenario: Selection follows the item across query edits
- **WHEN** an item is selected, and the user edits the filter query in a way that still matches that item
- **THEN** that same item remains selected at its new position in the narrowed list

#### Scenario: Selection is restored when the filter is cleared
- **WHEN** an item is selected under an active filter and the user presses `esc` to clear it
- **THEN** the full list is restored with that same item still selected

#### Scenario: Selection falls back when the item is filtered out
- **WHEN** the filter query is edited so that the previously selected item no longer matches
- **THEN** the selection falls back to the first visible row and the preview shows that row

### Requirement: Filter is scoped to one panel and visibly attributed
The filter SHALL apply to exactly one panel — the one focused when it was opened. Moving focus to another panel SHALL clear the filter, so that no panel is left silently narrowed. While a filter is active, the panel SHALL display the active query so the narrowed row set has a visible cause.

#### Scenario: Filter applies only to the focused panel
- **WHEN** the Specs panel is filtered
- **THEN** the Changes and Archive panels continue to show all of their items

#### Scenario: Switching panels clears the filter
- **WHEN** a filter is active on the Specs panel and the user presses `tab` to focus another panel
- **THEN** the filter is cleared and the Specs panel is restored to its full list

#### Scenario: The active query is shown
- **WHEN** a filter is active on a panel
- **THEN** that panel displays the query text, so the reduced number of rows is visibly attributable to the filter

### Requirement: Filter is unavailable from the preview
`/` SHALL open the list filter only while the list holds focus. While the preview pane holds focus, `/` SHALL open the preview search instead, leaving the list unfiltered.

#### Scenario: Preview focus routes slash to search
- **WHEN** the preview pane holds focus and the user presses `/`
- **THEN** the preview search opens and no list filter is applied
