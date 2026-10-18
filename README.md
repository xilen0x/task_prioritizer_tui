# tareas

A terminal task monitor written in Go with [Bubbletea](https://github.com/charmbracelet/bubbletea).
Up to four tasks on screen, one of them in focus, plus a pomodoro timer.
The UI is in Spanish and follows the `Tareas TUI` design (laid out for a 132×42 terminal).

## Run

```sh
go run ./cmd/tareas           # start (empty board on first run)
go run ./cmd/tareas --demo    # seed the sample tasks when no saved state exists
```

| Flag | Default | Meaning |
|------|---------|---------|
| `--bar=bloques\|ascii\|fino` | `bloques` | Progress bar glyphs |
| `--plain-digits` | `false` | Plain text instead of big block digits |
| `--demo` | `false` | Seed the four sample tasks if there is no saved state |

Quit with `F10` or `Ctrl+C`.

## Development

```sh
make check          # gofmt -l, go vet, go test
make update-golden  # regenerate TUI golden files
```
