// Command tareas is a terminal task monitor: up to four tasks on screen,
// one in focus, with a pomodoro timer.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/xilenox/tareas/internal/pomodoro"
	"github.com/xilenox/tareas/internal/storage/jsonfile"
	"github.com/xilenox/tareas/internal/tasks"
	"github.com/xilenox/tareas/internal/tui"
	"github.com/xilenox/tareas/internal/tui/theme"
)

// options holds the parsed command-line flags.
type options struct {
	bar         string
	plainDigits bool
	demo        bool
}

func parseFlags(args []string) (options, error) {
	var o options
	fs := flag.NewFlagSet("tareas", flag.ContinueOnError)
	fs.StringVar(&o.bar, "bar", "bloques", "progress bar style: bloques|ascii|fino")
	fs.BoolVar(&o.plainDigits, "plain-digits", false, "render big numbers as plain text")
	fs.BoolVar(&o.demo, "demo", false, "seed the sample tasks when no saved state exists")
	if err := fs.Parse(args); err != nil {
		return o, err
	}
	switch o.bar {
	case "bloques", "ascii", "fino":
	default:
		return o, fmt.Errorf("invalid --bar %q (want bloques|ascii|fino)", o.bar)
	}
	return o, nil
}

// initialState loads the saved state, or seeds a new one (the design's
// sample board with --demo, otherwise an empty board).
func initialState(store *jsonfile.Store, demo bool) (jsonfile.State, error) {
	st, found, err := store.Load()
	if err != nil {
		return st, err
	}
	if found {
		return st, nil
	}
	if demo {
		return jsonfile.State{
			Board: tasks.Demo(),
			Pomo:  pomodoro.Timer{Phase: pomodoro.Trabajo, Left: pomodoro.WorkSeconds - 437, Running: true, Done: 2},
		}, nil
	}
	return jsonfile.State{Board: tasks.New(), Pomo: pomodoro.New()}, nil
}

func run(o options) error {
	path, err := jsonfile.DefaultPath()
	if err != nil {
		return err
	}
	store := jsonfile.New(path)
	st, err := initialState(store, o.demo)
	if err != nil {
		return err
	}
	bar, _ := theme.ParseBarStyle(o.bar)
	m := tui.New(tui.Options{
		Board: st.Board,
		Pomo:  st.Pomo,
		Save: func(b tasks.Board, p pomodoro.Timer) error {
			return store.Save(jsonfile.State{Board: b, Pomo: p})
		},
		Bar:         bar,
		PlainDigits: o.plainDigits,
	})
	_, err = tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}

func main() {
	o, err := parseFlags(os.Args[1:])
	if err != nil {
		if !errors.Is(err, flag.ErrHelp) {
			fmt.Fprintln(os.Stderr, "tareas:", err)
		}
		os.Exit(2)
	}
	if err := run(o); err != nil {
		fmt.Fprintln(os.Stderr, "tareas:", err)
		os.Exit(1)
	}
}
