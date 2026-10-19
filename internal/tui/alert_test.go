package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/xilen0x/task_prioritizer_tui/internal/pomodoro"
)

type alertCall struct{ title, body string }

func tickWithAlert(t *testing.T, p pomodoro.Timer) []alertCall {
	t.Helper()
	var calls []alertCall
	m := New(Options{Pomo: p, NoTick: true, Alert: func(title, body string) {
		calls = append(calls, alertCall{title, body})
	}})
	_, cmd := m.Update(tickMsg{})
	runCmd(cmd)
	return calls
}

// runCmd executes cmd and, for a batch, every command inside it.
func runCmd(cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	if batch, ok := cmd().(tea.BatchMsg); ok {
		for _, c := range batch {
			runCmd(c)
		}
	}
}

func TestAlertOnPhaseChange(t *testing.T) {
	cases := []struct {
		name string
		pomo pomodoro.Timer
		want []alertCall
	}{
		{"work ends with a short break", pomodoro.Timer{Phase: pomodoro.Work, Left: 1, Running: true, Done: 0},
			[]alertCall{{"Pomodoro done", "Take a 5 min break."}}},
		{"fourth pomodoro gets a long break", pomodoro.Timer{Phase: pomodoro.Work, Left: 1, Running: true, Done: 3},
			[]alertCall{{"Pomodoro done", "Take a 15 min break."}}},
		{"break ends", pomodoro.Timer{Phase: pomodoro.Break, Left: 1, Running: true, Done: 1},
			[]alertCall{{"Break over", "Back to work: 25 min."}}},
		{"no alert mid-block", pomodoro.Timer{Phase: pomodoro.Work, Left: 100, Running: true}, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := tickWithAlert(t, c.pomo)
			if len(got) != len(c.want) {
				t.Fatalf("got %v, want %v", got, c.want)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Fatalf("got %v, want %v", got, c.want)
				}
			}
		})
	}
}

func TestNoAlertFuncIsSafe(t *testing.T) {
	m := New(Options{Pomo: pomodoro.Timer{Phase: pomodoro.Work, Left: 1, Running: true}, NoTick: true})
	_, cmd := m.Update(tickMsg{})
	runCmd(cmd)
}
