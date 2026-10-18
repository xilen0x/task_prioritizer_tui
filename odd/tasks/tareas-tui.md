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
- [ ] T1 Scaffold: go mod, deps (bubbletea, lipgloss, bubbles), main.go, Makefile, README — route: delegated (multi-file)
- [ ] T2 Domain `tasks` with table tests (TDD) — route: delegated
- [ ] T3 Domain `pomodoro` with tests (TDD) — route: delegated
- [ ] T4 Storage JSON adapter + round-trip tests — route: delegated
- [ ] T5 Theme + big-digit font + bar golden tests — route: delegated
- [ ] T6 TUI main screen + teatest golden at 132×42 — route: delegated
- [ ] T7 Overlays (modal/confirm/help/about) + key map tests — route: delegated
- [ ] T8 Resize degradation + best-effort mouse — route: delegated

## Acceptance criteria
- All behaviour rules and keys from the design hold (see plan "Source of truth").
- `go test ./...`, `go vet ./...` pass; `gofmt -l .` empty.
- `go run ./cmd/tareas --demo` renders the design's sample state.

## Progress / Evidence
(updated per task with commit hash and check output)

## Next step
T1.
