package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/muesli/termenv"

	"github.com/xilen0x/task_prioritizer_tui/internal/pomodoro"
	"github.com/xilen0x/task_prioritizer_tui/internal/tasks"
)

// Goldens are plain text: the ASCII profile drops colours and attributes.
func init() { lipgloss.SetColorProfile(termenv.Ascii) }

// fixedClock is Sunday 18 Oct 2026, 14:03:27.
func fixedClock() time.Time { return time.Date(2026, 10, 18, 14, 3, 27, 0, time.UTC) }

func demoPomo() pomodoro.Timer {
	return pomodoro.Timer{Phase: pomodoro.Work, Left: 1500 - 437, Running: true, Done: 2}
}

type saveSpy struct{ calls int }

func (s *saveSpy) save(tasks.Board, pomodoro.Timer) error { s.calls++; return nil }

func newDemo(spy *saveSpy) Model {
	o := Options{Board: tasks.Demo(), Pomo: demoPomo(), Clock: fixedClock, NoTick: true}
	if spy != nil {
		o.Save = spy.save
	}
	return New(o)
}

var namedKeys = map[string]tea.KeyType{
	"up": tea.KeyUp, "down": tea.KeyDown, "left": tea.KeyLeft, "right": tea.KeyRight,
	"enter": tea.KeyEnter, "esc": tea.KeyEsc, "tab": tea.KeyTab, "shift+tab": tea.KeyShiftTab,
	"delete": tea.KeyDelete, "backspace": tea.KeyBackspace,
	"f1": tea.KeyF1, "f2": tea.KeyF2, "f3": tea.KeyF3, "f4": tea.KeyF4, "f5": tea.KeyF5,
	"f6": tea.KeyF6, "f8": tea.KeyF8, "f9": tea.KeyF9, "f10": tea.KeyF10,
}

// key builds a KeyMsg from a name ("down", "f4", "space") or literal runes.
func key(s string) tea.KeyMsg {
	if s == "space" {
		return tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	}
	if t, ok := namedKeys[s]; ok {
		return tea.KeyMsg{Type: t}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func press(m Model, keys ...string) Model {
	for _, k := range keys {
		next, _ := m.Update(key(k))
		m = next.(Model)
	}
	return m
}

func requireScreen(t *testing.T, m Model, w, h int) {
	t.Helper()
	lines := strings.Split(m.View(), "\n")
	if len(lines) != h {
		t.Fatalf("view has %d lines, want %d", len(lines), h)
	}
	for i, l := range lines {
		if got := lipgloss.Width(l); got != w {
			t.Fatalf("line %d is %d wide, want %d: %q", i, got, w, l)
		}
	}
}

func TestMainScreenGolden(t *testing.T) {
	m := newDemo(nil)
	requireScreen(t, m, DesignWidth, DesignHeight)
	golden.RequireEqual(t, []byte(m.View()))
}

func TestEmptyScreenGolden(t *testing.T) {
	m := New(Options{Clock: fixedClock, NoTick: true})
	requireScreen(t, m, DesignWidth, DesignHeight)
	golden.RequireEqual(t, []byte(m.View()))
}

func TestPlainDigitsAndASCIIBarsGolden(t *testing.T) {
	m := New(Options{Board: tasks.Demo(), Pomo: demoPomo(), Clock: fixedClock, NoTick: true, Bar: "ascii", PlainDigits: true})
	requireScreen(t, m, DesignWidth, DesignHeight)
	golden.RequireEqual(t, []byte(m.View()))
}

func TestTeatestKeySequenceGolden(t *testing.T) {
	tm := teatest.NewTestModel(t, newDemo(nil), teatest.WithInitialTermSize(DesignWidth, DesignHeight))
	tm.Send(key("down"))  // select task 2
	tm.Send(key("enter")) // focus task 2
	tm.Send(key("down"))  // select task 1
	tm.Send(key("space")) // task 1 doing -> done
	tm.Send(key("p"))     // pause pomodoro
	tm.Send(key("f10"))
	final := tm.FinalModel(t, teatest.WithFinalTimeout(3*time.Second)).(Model)
	b := final.Board()
	if b.FocusID != 2 || b.SelID != 1 || b.Get(1).Status != tasks.Done || final.Pomo().Running {
		t.Fatalf("unexpected final state: focus=%d sel=%d status=%s running=%v", b.FocusID, b.SelID, b.Get(1).Status, final.Pomo().Running)
	}
	golden.RequireEqual(t, []byte(final.View()))
}

func TestMainKeys(t *testing.T) {
	tests := []struct {
		name  string
		keys  []string
		check func(t *testing.T, m Model)
	}{
		{"j and k move selection", []string{"j", "j", "k"}, func(t *testing.T, m Model) {
			if m.board.SelID != 2 {
				t.Fatalf("sel = %d", m.board.SelID)
			}
		}},
		{"up wraps to last", []string{"up"}, func(t *testing.T, m Model) {
			if m.board.SelID != 4 {
				t.Fatalf("sel = %d", m.board.SelID)
			}
		}},
		{"f5 focuses selection", []string{"down", "down", "f5"}, func(t *testing.T, m Model) {
			if m.board.FocusID != 3 {
				t.Fatalf("focus = %d", m.board.FocusID)
			}
		}},
		{"plus, equals and minus bump progress", []string{"+", "=", "-"}, func(t *testing.T, m Model) {
			if got := m.board.Get(1).Progress; got != 80 {
				t.Fatalf("progress = %d", got)
			}
		}},
		{"f4 cycles status", []string{"f4"}, func(t *testing.T, m Model) {
			if got := m.board.Get(1).Status; got != tasks.Done {
				t.Fatalf("status = %s", got)
			}
		}},
		{"r resets and pauses the pomodoro", []string{"r"}, func(t *testing.T, m Model) {
			if m.pomo.Running || m.pomo.Left != pomodoro.WorkSeconds || m.pomo.Done != 2 {
				t.Fatalf("pomo = %+v", m.pomo)
			}
		}},
		{"f6 toggles the pomodoro", []string{"f6"}, func(t *testing.T, m Model) {
			if m.pomo.Running {
				t.Fatal("still running")
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spy := &saveSpy{}
			m := press(newDemo(spy), tt.keys...)
			tt.check(t, m)
			if spy.calls != len(tt.keys) {
				t.Fatalf("saved %d times, want %d", spy.calls, len(tt.keys))
			}
		})
	}
}

func TestQuitSavesAndQuits(t *testing.T) {
	for _, k := range []tea.KeyMsg{key("f10"), {Type: tea.KeyCtrlC}} {
		spy := &saveSpy{}
		_, cmd := newDemo(spy).Update(k)
		if cmd == nil || spy.calls != 1 {
			t.Fatalf("%s: cmd=%v saves=%d", k, cmd, spy.calls)
		}
		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Fatalf("%s did not quit", k)
		}
	}
}

func TestTickAdvancesPomodoroAndFocusTime(t *testing.T) {
	m := newDemo(nil)
	next, _ := m.Update(tickMsg{})
	m = next.(Model)
	if m.pomo.Left != 1062 || m.board.Get(1).Elapsed != 5048 {
		t.Fatalf("left=%d elapsed=%d", m.pomo.Left, m.board.Get(1).Elapsed)
	}
	m = press(m, "p")
	next, _ = m.Update(tickMsg{})
	m = next.(Model)
	if m.pomo.Left != 1062 || m.board.Get(1).Elapsed != 5048 {
		t.Fatal("paused pomodoro must not advance")
	}
}

func TestTickRefreshesClock(t *testing.T) {
	now := fixedClock()
	m := New(Options{Clock: func() time.Time { return now }, NoTick: true})
	now = now.Add(time.Second)
	next, _ := m.Update(tickMsg{})
	if !strings.Contains(next.(Model).View(), "14:03:28") {
		t.Fatal("clock not refreshed on tick")
	}
}
