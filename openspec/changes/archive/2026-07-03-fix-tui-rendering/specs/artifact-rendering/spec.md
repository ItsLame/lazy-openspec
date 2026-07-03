## MODIFIED Requirements

### Requirement: Theme adaptation
The application SHALL choose a rendering theme appropriate to the terminal (light/dark) and SHALL degrade gracefully when the terminal lacks colour support. The colours used for OpenSpec's semantically rendered structures — requirement headers, `WHEN`/`THEN`/`AND` keywords, delta-operation badges, progress bars, and checklist glyphs — SHALL be drawn from the terminal's ANSI 16-colour palette (indices 0–15) rather than fixed xterm-256 indices, so that the main pane matches the panel chrome and the user's terminal theme.

#### Scenario: No-color terminal
- **WHEN** the terminal does not support colour (e.g. `NO_COLOR` is set or the output is not a TTY-capable color terminal)
- **THEN** artifacts render in a monochrome style that remains readable without relying on colour to convey structure

#### Scenario: Semantic colours match the terminal theme
- **WHEN** requirements, scenarios, operation badges, or task checklists are rendered in a colour-capable terminal
- **THEN** their accent colours (e.g. green for `WHEN`/added, red for removed, yellow for active/modified) come from the terminal's ANSI palette and therefore match the same theme colours used by the surrounding panel borders and selection highlight
