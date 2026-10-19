# Feature: big-titles-english

## Objective
1. Translate the whole UI to English.
2. Render task titles roughly twice as tall using a block-glyph font (works in VTE terminals and Ghostty, where native DECDHL double-height lines are not supported).

## Why
User request. Done on branch `feat/big-titles-english` so it can be discarded if the user does not like it. The English work and the big titles are separate commits, so English can be kept alone.

## Scope
- Every user-visible string goes to English: header, labels, statuses, priorities, hints, F-key bar, modal, confirm, help, about, small-terminal notice, save error, dates.
- The persisted JSON stays readable: old Spanish status values still load.
- A block font for titles (uppercase A–Z, 0–9, common punctuation, accents folded to base letters), 3 rows tall. The focus title gets up to 2 lines; each row title gets 1 line, with an ellipsis when it overflows. Compact layouts fall back to plain bold.

## Tasks
- [x] T1 English UI + status value migration + tests/goldens — route: delegated (multi-file)
- [ ] T2 Block-font titles + layout fit at 132×42 + goldens — route: delegated (multi-file)

## Acceptance criteria
- No Spanish strings left in the UI. Old state files still load.
- Titles render in the block font at 132×42 and every screen stays exactly W×H.
- `gofmt -l .` is empty, and `go vet ./...` and `go test ./...` pass.

## Progress / Evidence
### T1 (done)
- Every UI string is in English, dates use English abbreviations ("Sun 18 Oct") and hours use a decimal dot ("1.5 h").
- Domain values are renamed: status `todo/doing/done` (`tasks.Todo/Doing/Done`), pomodoro phase `work/break` (`pomodoro.Work/Break`) and bar style `blocks/ascii/thin` (`theme.Blocks/Thin`).
- Storage maps the legacy values (`pendiente/curso/hecha`, `trabajo/descanso`) on load; the `--bar` flag still accepts `bloques/fino`. The binary path `cmd/tareas` and the data path `~/.local/share/tareas/state.json` are unchanged.
- The delete confirm now uses y/Y/Enter, and n/N/Esc cancel. `s` no longer deletes.
- Test-first: `TestLoadMapsLegacySpanishValues` and the y/Y cases of `TestConfirmKeys` were observed RED before the change and GREEN after it.
- Checks: `gofmt -l .` printed nothing; `go vet ./...` passed; `go test -count=1 ./...` passed (6 packages); `go build ./cmd/tareas` passed. Goldens were regenerated and inspected.

## Next step
T2.
