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

## Keys

| Key | Action |
|-----|--------|
| `↑` `↓` / `j` `k` | Move the selection |
| `Enter` / `F5` | Put the selected task in focus |
| `n` / `F2` | New task (at most 4) |
| `e` / `F3` | Edit the selected task |
| `Space` / `F4` | Cycle status: pendiente → en curso → hecha |
| `+` `=` / `-` | Progress ±10% |
| `x` / `Del` / `F8` | Delete (confirm with `s`/`Enter`, cancel with `n`/`Esc`) |
| `p` / `F6` | Pause or resume the pomodoro |
| `r` | Reset the pomodoro |
| `F1` / `?` | Help |
| `a` / `F9` | About |

In the editor: `Tab`/`Shift+Tab` or `↑`/`↓` change field, `←`/`→` change the value,
`Enter` saves and `Esc` cancels. The pomodoro and tracked time pause while the editor is open.
Task time only accrues for the focused task, while it is *en curso* and the pomodoro is in a running work block.

Mouse support is best effort: click a row, the focus box or a sidebar column to select,
click a free slot to create a task, click the function-key bar, scroll to move the selection.

## Data

State is saved as JSON to `$XDG_DATA_HOME/tareas/state.json`
(default `~/.local/share/tareas/state.json`) after every change, every few seconds while the
pomodoro runs, and on quit. Writes are atomic (temp file + rename).

## Terminal size

The layout targets 132×42. Smaller terminals get a compact layout with plain digits;
below 100×30 a "terminal too small" notice is shown.

## Development

```sh
make check          # gofmt -l, go vet, go test
make update-golden  # regenerate TUI golden files
```
