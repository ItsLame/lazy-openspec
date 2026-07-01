## ADDED Requirements

### Requirement: Readable prose rendering
The application SHALL render free-text markdown sections of artifacts (e.g. Why, Context, Decisions, Risks) using a terminal markdown renderer (Glamour) so that headings, emphasis, lists, blockquotes, and fenced code are styled and word-wrapped to the pane width, rather than shown as raw markdown source.

#### Scenario: Render proposal prose
- **WHEN** a proposal is shown in the main pane
- **THEN** its headings are styled (without literal `#` characters), bullet lists use glyphs, and long lines are wrapped to the current pane width with hanging indents

#### Scenario: Narrow pane rewrap
- **WHEN** the main pane width changes due to a resize or panel layout change
- **THEN** the rendered markdown re-wraps to the new width without cutting words mid-glyph

### Requirement: Semantic rendering of requirements and scenarios
The application SHALL render OpenSpec's known structures with dedicated styling instead of generic markdown — requirement headers as labelled/badged headings, and scenarios with their `WHEN`/`THEN` keywords visually emphasized and aligned.

#### Scenario: WHEN/THEN emphasis
- **WHEN** a scenario is rendered
- **THEN** its `WHEN` and `THEN` keywords are visually distinct (e.g. coloured and aligned) and the condition/outcome text follows on the same visual row

### Requirement: Semantic rendering of task checklists
The application SHALL render `tasks.md` as a grouped checklist, showing task groups with per-group progress and each task with a checkbox glyph (checked/unchecked), dimming completed tasks.

#### Scenario: Render task groups
- **WHEN** the tasks artifact is shown
- **THEN** each numbered task group displays a progress indicator and each task shows a `☐` or `✔` glyph, with completed tasks rendered dimmed

### Requirement: Theme adaptation
The application SHALL choose a rendering theme appropriate to the terminal (light/dark) and SHALL degrade gracefully when the terminal lacks colour support.

#### Scenario: No-color terminal
- **WHEN** the terminal does not support colour (e.g. `NO_COLOR` is set or the output is not a TTY-capable color terminal)
- **THEN** artifacts render in a monochrome style that remains readable without relying on colour to convey structure
