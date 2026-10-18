// Command tareas is a terminal task monitor: up to four tasks on screen,
// one in focus, with a pomodoro timer.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
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

// placeholder is a minimal model used until the real UI is wired in.
type placeholder struct{}

func (placeholder) Init() tea.Cmd { return nil }

func (p placeholder) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok && (k.Type == tea.KeyF10 || k.Type == tea.KeyCtrlC) {
		return p, tea.Quit
	}
	return p, nil
}

func (placeholder) View() string { return "tareas — F10 para salir\n" }

func main() {
	if _, err := parseFlags(os.Args[1:]); err != nil {
		if !errors.Is(err, flag.ErrHelp) {
			fmt.Fprintln(os.Stderr, "tareas:", err)
		}
		os.Exit(2)
	}
	if _, err := tea.NewProgram(placeholder{}, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "tareas:", err)
		os.Exit(1)
	}
}
