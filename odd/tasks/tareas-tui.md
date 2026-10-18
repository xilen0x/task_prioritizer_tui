# Feature: tareas-tui

## Objective
Implement the Claude Design mock-up `Tareas TUI.dc.html` (project 8b6fa1df-871b-46bc-9050-1d86ee9958b2) as a real terminal app in Go + Bubbletea.

## Problem / Why
The design exists only as a browser mock-up; the user wants a working terminal task monitor (max 4 tasks, one in focus, pomodoro timer).

## Scope
- Domain: tasks board, pomodoro timer (pure Go, no TUI imports).
- JSON persistence adapter ($XDG_DATA_HOME/tareas/state.json), `--demo` seed.
- Bubbletea UI faithful to the design: layout for 132×42, Spanish UI copy, design colours, keys, overlays.
- Flags: `--bar=bloques|ascii|fino`, `--plain-digits`, `--demo`.

## Constraints
- Hexagonal/screaming layout: `internal/tasks`, `internal/pomodoro`, `internal/storage/jsonfile`, `internal/tui`, `internal/tui/theme`, `cmd/tareas`.
- Conventional commits, no AI attribution.
- Code/comments/docs in English; UI copy in Spanish (matches design).

## Delivery strategy
`ask-on-risk`. Forecast > 400 authored lines; feature is local-only (no remote yet), so PR slicing is deferred until the user adds a remote.

## Tasks
- [x] T1 Scaffold: go mod, deps (bubbletea, lipgloss, bubbles), main.go, Makefile, README — route: delegated (multi-file)
- [x] T2 Domain `tasks` with table tests (TDD) — route: delegated
- [x] T3 Domain `pomodoro` with tests (TDD) — route: delegated
- [x] T4 Storage JSON adapter + round-trip tests — route: delegated
- [x] T5 Theme + big-digit font + bar golden tests — route: delegated
- [x] T6 TUI main screen + teatest golden at 132×42 — route: delegated
- [x] T7 Overlays (modal/confirm/help/about) + key map tests — route: delegated
- [x] T8 Resize degradation + best-effort mouse — route: delegated

## Acceptance criteria
- All behaviour rules and keys from the design hold (see plan "Source of truth").
- `go test ./...`, `go vet ./...` pass; `gofmt -l .` empty.
- `go run ./cmd/tareas --demo` renders the design's sample state.

## Progress / Evidence
(updated per task with commit hash and check output)

- T1 `5d8e498` chore: scaffold. gofmt clean, go vet ok, go build ok, `--bar=x` exits 2 with message.
- T2 `e967028` feat(tasks): domain. RED: tests failed to compile (undefined Task/Demo/New); GREEN: `go test ./internal/tasks` ok; vet/gofmt clean.
- T3 `607da86` feat(pomodoro): pure timer. RED: undefined New/Timer; GREEN: `go test ./internal/pomodoro` ok (phase transitions, long break every 4, full 1500-tick block).
- T4 `47fc3ba` feat(storage): JSON adapter. RED: undefined New/State; GREEN: round-trip, missing, corrupt, repair, XDG/HOME fallback tests ok with t.TempDir.
- T5 `8df022a` feat(theme): palette, big-digit font, bars, formatters. RED: undefined Big; GREEN after `-update`: big/bars goldens, widths, 62-cell rounding, mix() reproduces on-dark hex levels.
- T6 `72b68c4` feat(tui): main screen + main/ wiring to JSON store. Goldens (ASCII profile, fixed clock, NoTick): demo, empty, plain-digits+ascii bars, teatest key sequence (down/enter/down/space/p/F10) final-model view; every line asserted 132 wide × 42 lines. Key table tests assert one save per mutation; F10/ctrl+c save+quit. main tests: flags, --demo seed, saved state wins.
- T7 `602d03c` feat(tui): edit/new modal (bubbles textinput, static cursor), delete confirm, help, about, composited over a dimmed scrim. Tests written alongside the implementation (no separate RED run): editor save/nav/esc, empty title, 4-task limit, modal pauses ticks (help does not), confirm s/S/Enter/n/N/Esc, help/about close keys; 5 overlay goldens; TrueColor geometry check. Board read methods switched to value receivers.
- T8 feat(tui): computed layout (compact <132×42 forces plain digits, tighter rows, shorter sidebar bars, compact key bar; <100×30 shows notice; larger sizes grow), best-effort mouse (row/focus/sidebar select, free slot opens editor, F-key bar, wheel, click closes help/about) via tea.WithMouseCellMotion. Resize goldens 100×30, 120×36, 150×50, 80×24; mouse table tests. README documents keys, data path and sizes. Final: gofmt -l empty, go vet ok, go test ./... ok, go build ./cmd/tareas ok.

## Next step
All tasks done. Manual check in a real 132×42 terminal (`go run ./cmd/tareas --demo`) against the design preview; then the user decides on push/PR.

## Known gaps
- No double-click-to-edit (design onDoubleClick) - Bubbletea v1 has no double-click event.
- Runtime visual check in a real terminal was not run by the writer (no tmux in the sandbox); goldens are ASCII-profile plus a TrueColor geometry test.
