// Command task_prioritizer is a terminal task monitor: up to four tasks on screen,
// one in focus, with a pomodoro timer.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/xilen0x/task_prioritizer_tui/internal/notify"
	"github.com/xilen0x/task_prioritizer_tui/internal/pomodoro"
	"github.com/xilen0x/task_prioritizer_tui/internal/storage/jsonfile"
	"github.com/xilen0x/task_prioritizer_tui/internal/tasks"
	"github.com/xilen0x/task_prioritizer_tui/internal/tui"
	"github.com/xilen0x/task_prioritizer_tui/internal/tui/theme"
)

// options holds the parsed command-line flags.
type options struct {
	bar         string
	plainDigits bool
	demo        bool
	noAlert     bool
}

func parseFlags(args []string) (options, error) {
	var o options
	fs := flag.NewFlagSet("task_prioritizer", flag.ContinueOnError)
	fs.StringVar(&o.bar, "bar", "blocks", "progress bar style: blocks|ascii|thin")
	fs.BoolVar(&o.plainDigits, "plain-digits", false, "render big numbers as plain text")
	fs.BoolVar(&o.demo, "demo", false, "seed the sample tasks when no saved state exists")
	fs.BoolVar(&o.noAlert, "no-alert", false, "no bell or desktop notification when a pomodoro phase ends")
	if err := fs.Parse(args); err != nil {
		return o, err
	}
	if _, ok := theme.ParseBarStyle(o.bar); !ok {
		return o, fmt.Errorf("invalid --bar %q (want blocks|ascii|thin)", o.bar)
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
			Pomo:  pomodoro.Timer{Phase: pomodoro.Work, Left: pomodoro.WorkSeconds - 437, Running: true, Done: 2},
		}, nil
	}
	return jsonfile.State{Board: tasks.New(), Pomo: pomodoro.New()}, nil
}

func run(o options) error {
	path, err := jsonfile.DefaultPath()
	if err != nil {
		return err
	}
	if err := jsonfile.MigrateLegacy(path); err != nil {
		return err
	}
	store := jsonfile.New(path)
	st, err := initialState(store, o.demo)
	if err != nil {
		return err
	}
	bar, _ := theme.ParseBarStyle(o.bar)
	var alert tui.AlertFunc
	if !o.noAlert {
		alert = notify.New().Alert
	}
	m := tui.New(tui.Options{
		Board: st.Board,
		Pomo:  st.Pomo,
		Save: func(b tasks.Board, p pomodoro.Timer) error {
			return store.Save(jsonfile.State{Board: b, Pomo: p})
		},
		Bar:         bar,
		PlainDigits: o.plainDigits,
		Alert:       alert,
	})
	_, err = tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run()
	return err
}

func main() {
	o, err := parseFlags(os.Args[1:])
	if err != nil {
		if !errors.Is(err, flag.ErrHelp) {
			fmt.Fprintln(os.Stderr, "task_prioritizer:", err)
		}
		os.Exit(2)
	}
	if err := run(o); err != nil {
		fmt.Fprintln(os.Stderr, "task_prioritizer:", err)
		os.Exit(1)
	}
}
