## Context

OpenSpec is a Node/TypeScript CLI for spec-driven development. Its only visual surface is `openspec view`, a one-shot static printout (chalk, ~167 lines). Everything interactive today is done through discrete commands (`list`, `show`, `status`, `validate`, `apply`, `archive`) and by opening markdown artifacts in an editor.

`lazy-openspec` is a new standalone terminal application, built in this `lazy-openspec` repository, that turns the whole loop into one `lazygit`/`lazydocker`-style screen. It is a *client* of the OpenSpec CLI: it does not link OpenSpec code, it shells out to `openspec … --json`. The important constraints:

- The CLI already exposes structured JSON for everything the UI needs: `list --json`, `list --specs --json`, `show --json` (with `--requirements`, `--no-scenarios`, `-r <id>`), `status --change <name> --json`, `instructions --json`. This is treated as the integration contract.
- The change lifecycle (draft → active → completed → archived) and task progress are already computed by the CLI from `tasks.md` checkbox counts.
- Artifacts are markdown with known structure (proposal Why/What/Capabilities/Impact; spec Requirement/Scenario/WHEN/THEN; tasks numbered `- [ ]` groups; design prose).

## Goals / Non-Goals

**Goals:**
- A single, keyboard-driven, `lazygit`-style TUI: stacked numbered panels (Changes, Specs, Archive), a main content pane, and a bottom command-log/hint bar.
- Comfortable reading of artifacts: Glamour for prose plus semantic rendering of OpenSpec blocks.
- Interactive task toggling that persists to `tasks.md`.
- Running `validate` / `apply` / `archive` from the UI with streamed output.
- Ship as a single static Go binary with no runtime dependencies other than the `openspec` CLI itself.
- Resilience: never crash on a missing CLI, missing root, malformed JSON, or a resize.

**Non-Goals:**
- Reimplementing OpenSpec parsing, the lifecycle model, or the schema engine — all of that stays in the CLI.
- Authoring artifacts via rich in-app forms (creating a change, writing prose). v1 supports the `proposal → tasks` review/execution loop and delegates content authoring to `$EDITOR` / the existing `/opsx` skills.
- Upstreaming into `openspec` itself, or supporting non-`spec-driven` schemas, in v1.
- A general markdown editor. The only in-place write v1 performs is toggling task checkboxes.

## Decisions

### D1: Go + Bubble Tea + Lip Gloss + Glamour
Chosen over TypeScript/Ink and Rust/ratatui. Bubble Tea gives a mature Elm-style (Model/Update/View) loop that fits multi-panel navigation; Lip Gloss handles borders/layout/styling; **Glamour** renders markdown beautifully out of the box, which directly answers the "make markdowns easily readable" requirement. The whole thing compiles to one static binary.
- *Alternative — TS/Ink:* same ecosystem as OpenSpec and could import internals, but terminal markdown rendering is DIY (`marked` + `marked-terminal` + `cli-highlight`) and needs a Node runtime. Rejected because the marquee feature (readable markdown) is weaker.
- *Alternative — Rust/ratatui:* fastest and robust, but the largest build and furthest from OpenSpec's ecosystem, with markdown rendering via `termimad`/custom. Rejected on effort.

### D2: Data via `openspec … --json` subprocess, not internals or raw files
The UI spawns the CLI and parses JSON. This keeps a stable public contract, works from Go, and reuses OpenSpec's own parsing and lifecycle logic (no drift).
- *Alternative — parse `openspec/` markdown directly:* zero CLI dependency and enables file-watching, but re-implements parsing and status semantics and risks drifting from OpenSpec. Rejected for v1; may be added later as a fast-path/live-reload optimization behind the same data interface.
- *Mitigation for subprocess cost:* an in-memory cache keyed by (command, args); a subprocess runs only on first load, explicit refresh (`r`), or after a mutation. Subprocess calls run as async Bubble Tea `Cmd`s so the UI never blocks.

### D3: lazygit-style layout and interaction model
Left column = stacked panels each with a number (`1` Changes, `2` Specs, `3` Archive); `tab`/`shift+tab` cycle focus; the focused panel gets a highlighted border. The right main pane shows a live preview of the current selection and, on `enter`, a drill-in detail view with an artifact tab bar. A bottom bar doubles as a command log (last command + streamed output) and a context-sensitive keybinding hint line. `?` opens a help overlay; `x` opens a context menu of actions for the selection.
- *Alternative — single-list master/detail (k9s-style):* simpler, but the user explicitly asked for the lazy\* multi-panel feel. Adopted lazy\* conventions deliberately (number keys, `x` menu, command log, `?` help).

### D4: Hybrid artifact rendering
Prose sections render through Glamour with a theme chosen from terminal background/`NO_COLOR`. OpenSpec's structured blocks are rendered semantically from `show --json` rather than as generic markdown: requirements as badged headers, scenarios with aligned/coloured `WHEN`/`THEN`, tasks as a grouped checklist with per-group progress and `☐`/`✔` glyphs (completed dimmed). This gives both correctness (structured data) and readability (Glamour for the free text).
- *Rationale:* rendering `#### Scenario` / `- **WHEN**` as raw markdown is noisy; the CLI hands us the parsed structure, so we style it directly.

### D5: Task toggling as a minimal, surgical file edit
Toggling rewrites only the single checkbox marker (`- [ ]` ⇄ `- [x]`) on the matching line in `tasks.md`, preserving every other byte, then re-reads status. This is done directly on the file (fast, precise) rather than through a CLI command, because OpenSpec has no "set task state" command. All other mutations (`validate`/`apply`/`archive`) go through the CLI.
- *Risk noted below:* writing the file directly must be robust to concurrent external edits.

### D6: Architecture / package layout
Three layers behind clear interfaces so the data source and renderer can evolve independently:
- `openspec` client package: typed structs + functions wrapping each `openspec … --json` call; owns caching and the `--store` flag.
- `render` package: Glamour setup, theme selection, and the semantic renderers for requirements/scenarios/tasks.
- `tui` package: Bubble Tea `Model`s (app root + per-panel + detail view), `Update` (key handling, async `Cmd`s), and `View` (Lip Gloss composition). A root model owns focus state and dispatches messages.

## Risks / Trade-offs

- **CLI JSON contract drift** → Pin/annotate the OpenSpec versions the client is tested against; centralize all parsing in the `openspec` client package with tolerant decoding (ignore unknown fields) so additive CLI changes don't break the UI; surface a clear message on decode failure instead of crashing.
- **Subprocess latency/flicker on large repos** → Cache aggressively, run CLI calls as async `Cmd`s with a spinner, and only refresh the affected slice after a mutation rather than reloading everything.
- **Direct `tasks.md` edit races an external editor** → Re-read the file immediately before toggling and match the target line by its task number + text (not by cached line index); if the expected marker isn't found, abort the toggle and show a "file changed on disk, refresh" message.
- **Destructive `archive` run by accident** → Require an explicit confirmation prompt (D3's `x` menu / dedicated `A` binding) before archiving.
- **Terminal/color capability variance** → Detect `NO_COLOR`/non-color terminals and fall back to a monochrome, glyph-and-layout-based style that stays readable; guard against tiny terminal sizes with a minimum-size message.
- **Scope of "full" v1** → Running `apply` from the UI can be long-running and interactive; v1 streams output read-only and does not attempt to proxy interactive prompts. If a command needs interactivity, the UI reports that and suggests running it in a shell.

## Migration Plan

This is greenfield and additive; there is nothing to migrate and no rollback surface in OpenSpec. Rollout is simply: build and distribute the `lazy-openspec` binary. Suggested delivery order — (1) Go module + CLI client + Bubble Tea skeleton, (2) read-only browsing + Glamour rendering, (3) semantic renderers, (4) task toggling, (5) workflow actions + command log. Each stage is independently demoable, so the tool is useful before the write features land.

## Open Questions

- Should v1 support schemas other than `spec-driven` (which fixes the artifact tab set), or hard-code the four tabs and generalize later?
- Distribution: `go install`, prebuilt release binaries, and/or Homebrew — which for the first release?
- Should file-watching (live reload of the `openspec/` tree) be in v1, or deferred until the subprocess-cache model proves insufficient?
- Command name/aliases: is `lazy-openspec` the binary name, with `lazyopsx`/`lopsx` as short aliases?
