# lazy-openspec

A [lazygit](https://github.com/jesseduffield/lazygit)/[lazydocker](https://github.com/jesseduffield/lazydocker)-style terminal UI for [OpenSpec](https://github.com/Fission-AI/OpenSpec). Browse changes and specs, read artifacts rendered beautifully, tick off tasks, and run the OpenSpec workflow commands — all from one keyboard-driven screen.

```
┌─1 Changes────────────┐┌─ add-user-auth ─────────────── active · 3/5 ─┐
│ Active               ││ proposal · specs · design · tasks            │
│▸◉ add-user-auth  60% ││                                              │
│ Draft                ││ 1. Backend auth              ██████░ 2/3     │
│ ○ add-data-export    ││   ✔ 1.1 Add user model                       │
├─2 Specs──────────────┤│ ▸ ☐ 1.3 Issue session tokens                 │
│ ▪ auth        4r     ││                                              │
├─3 Archive────────────┤│ 2. Frontend                  ░░░░░░░ 0/2     │
│ ▫ old-migration      ││   ☐ 2.1 Build login form      scroll 40% ────│
└──────────────────────┘└──────────────────────────────────────────────┘
┌─ Command log ────────────────────────────────────────────────────────┐
│ $ openspec validate add-user-auth → ✓ completed                      │
└──────────────────────────────────────────────────────────────────────┘
 [ ] artifact  space toggle  v validate  esc back  ? help
```

## Why

OpenSpec's built-in `openspec view` is a one-shot printout. `lazy-openspec` is a
full interactive TUI: stacked, numbered panels on the left (Changes, Specs,
Archive), a rendered detail pane on the right, and a command-log/hint bar at the
bottom. Artifacts are rendered readably — [Glamour](https://github.com/charmbracelet/glamour)
for prose, and semantic rendering of OpenSpec's structures (requirements,
`WHEN`/`THEN` scenarios, task checklists).

## Install

Requires the [`openspec`](https://github.com/Fission-AI/OpenSpec) CLI on your
`PATH` — `lazy-openspec` shells out to it.

```sh
go install github.com/itslame/lazy-openspec/cmd/lazy-openspec@latest
# or from a clone:
make install
```

## Usage

Run it inside a directory that has an `openspec/` root:

```sh
lazy-openspec
lazy-openspec --store <id>   # target a registered OpenSpec store
```

## Keybindings

| Key | Action |
| --- | --- |
| `tab` / `shift+tab`, `1`–`3` | switch / jump between panels |
| `↑`/`↓`, `j`/`k` | move selection (or scroll) |
| `enter` | open the selected change or spec |
| `[` / `]`, `←`/`→` | switch artifact tab (proposal · specs · design · tasks) |
| `space` | toggle the selected task (tasks tab), persisted to `tasks.md` |
| `n` / `p` | next / previous requirement (spec view) |
| `v` | run `openspec validate` |
| `a` | show `openspec` apply instructions |
| `A` | run `openspec archive` (with confirmation) |
| `x` | actions menu |
| `r` | refresh |
| `?` | help overlay |
| `esc` | back |
| `q` / `ctrl+c` | quit |

## Architecture

- `internal/openspec` — data-access layer: shells out to `openspec … --json`
  (`list`, `status`, `show`, `spec show`), decodes tolerantly, caches results.
- `internal/render` — Glamour prose rendering + semantic renderers for
  requirements, scenarios, and task checklists (with a monochrome `NO_COLOR`
  fallback).
- `internal/tasks` — parses `tasks.md` and performs the byte-preserving checkbox
  toggle.
- `internal/tui` — the [Bubble Tea](https://github.com/charmbracelet/bubbletea)
  models, [Lip Gloss](https://github.com/charmbracelet/lipgloss) layout, and the
  streaming command runner.

## Notes

- **`a` shows apply instructions** rather than "running apply": there is no
  `openspec apply` CLI command — apply is an AI-driven step (e.g. the `/opsx:apply`
  skill). The `a` key surfaces `openspec instructions apply`, the real affordance.
- Tested against `openspec` (@fission-ai/openspec) **1.5.0**.

## Development

```sh
make test    # unit + (skippable) live integration tests
make vet
make build   # -> bin/lazy-openspec
```

Built with Go, Bubble Tea, Lip Gloss, and Glamour.
