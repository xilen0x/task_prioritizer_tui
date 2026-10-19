package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/muesli/termenv"

	"github.com/xilen0x/task_prioritizer_tui/internal/tasks"
)

func TestOverlayGoldens(t *testing.T) {
	tests := []struct {
		name string
		keys []string
	}{
		{"edit", []string{"e", "down", "down", "right"}},
		{"new", []string{"down", "x", "y", "n", "N", "e", "w"}},
		{"confirm", []string{"down", "x"}},
		{"help", []string{"?"}},
		{"about", []string{"a"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := press(newDemo(nil), tt.keys...)
			requireScreen(t, m, DesignWidth, DesignHeight)
			golden.RequireEqual(t, []byte(m.View()))
		})
	}
}

func TestEditorSavesDraft(t *testing.T) {
	spy := &saveSpy{}
	// select task 2, edit: retitle, status → done, prio → P1, estimate +1h
	m := press(newDemo(spy), "down", "e", "backspace", "backspace", "X",
		"tab", "right", "tab", "left", "tab", "tab", "right", "right", "enter")
	got := m.board.Get(2)
	want := tasks.Task{ID: 2, Title: "Review the backups module X", Status: tasks.Done, Prio: 1, Progress: 100, Est: 3, Elapsed: 1820}
	if got != want {
		t.Fatalf("got  %+v\nwant %+v", got, want)
	}
	if m.editor != nil || spy.calls != 2 { // one save for "down", one for "enter"
		t.Fatalf("editor open=%v saves=%d", m.editor != nil, spy.calls)
	}
}

func TestEditorNavigation(t *testing.T) {
	m := press(newDemo(nil), "e")
	steps := []struct {
		key  string
		want int
	}{
		{"up", fieldTitle}, {"down", fieldStatus}, {"down", fieldPrio}, {"down", fieldProgress},
		{"down", fieldEst}, {"down", fieldEst}, {"tab", fieldTitle}, {"shift+tab", fieldEst},
	}
	for _, s := range steps {
		m = press(m, s.key)
		if m.editor.field != s.want {
			t.Fatalf("after %s field = %d, want %d", s.key, m.editor.field, s.want)
		}
	}
}

func TestEditorTitleIgnoresArrowsForValues(t *testing.T) {
	m := press(newDemo(nil), "e", "left", "right")
	if m.editor.draft != m.board.Get(1) {
		t.Fatal("left/right on the title field must not change values")
	}
}

func TestEditorEscCancels(t *testing.T) {
	spy := &saveSpy{}
	m := press(newDemo(spy), "e", "X", "tab", "right", "esc")
	if m.editor != nil || m.board.Get(1) != tasks.Demo().Get(1) || spy.calls != 0 {
		t.Fatalf("esc must discard the draft: %+v saves=%d", m.board.Get(1), spy.calls)
	}
}

func TestNewTaskEmptyTitleAndLimit(t *testing.T) {
	m := New(Options{Clock: fixedClock, NoTick: true})
	m = press(m, "f2", "enter")
	if len(m.board.Tasks) != 1 || m.board.Tasks[0].Title != tasks.UntitledName || m.board.FocusID != 1 {
		t.Fatalf("got %+v", m.board)
	}
	full := press(newDemo(nil), "n")
	if full.editor != nil {
		t.Fatal("n must do nothing with 4 tasks")
	}
	if !strings.Contains(full.View(), "F2  New") {
		t.Fatal("key bar must still list New")
	}
}

func TestModalPausesTicks(t *testing.T) {
	m := press(newDemo(nil), "e")
	next, _ := m.Update(tickMsg{})
	m = next.(Model)
	if m.pomo.Left != 1063 || m.board.Get(1).Elapsed != 5047 {
		t.Fatalf("modal open: left=%d elapsed=%d", m.pomo.Left, m.board.Get(1).Elapsed)
	}
	m = press(m, "esc", "?")
	next, _ = m.Update(tickMsg{})
	if next.(Model).pomo.Left != 1062 {
		t.Fatal("help must not pause the pomodoro")
	}
}

func TestConfirmKeys(t *testing.T) {
	tests := []struct {
		name    string
		keys    []string
		wantLen int
	}{
		{"y deletes", []string{"x", "y"}, 3},
		{"Y deletes", []string{"x", "Y"}, 3},
		{"s no longer deletes", []string{"x", "s"}, 4},
		{"enter deletes", []string{"delete", "enter"}, 3},
		{"f8 then n cancels", []string{"f8", "n"}, 4},
		{"N cancels", []string{"x", "N"}, 4},
		{"esc cancels", []string{"x", "esc"}, 4},
		{"other keys keep dialog", []string{"x", "q"}, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := press(newDemo(nil), tt.keys...)
			if len(m.board.Tasks) != tt.wantLen {
				t.Fatalf("len = %d, want %d", len(m.board.Tasks), tt.wantLen)
			}
		})
	}
	m := press(newDemo(nil), "x", "y")
	if m.board.FocusID != 2 || m.board.SelID != 2 {
		t.Fatalf("deleting focus: focus=%d sel=%d", m.board.FocusID, m.board.SelID)
	}
}

func TestHelpAndAboutClose(t *testing.T) {
	for _, open := range []string{"?", "f1", "a", "f9"} {
		for _, closeKey := range []string{"esc", "f1", "?", "a", "f9"} {
			m := press(newDemo(nil), open)
			if !m.help && !m.about {
				t.Fatalf("%s did not open", open)
			}
			m = press(m, closeKey)
			if m.help || m.about {
				t.Fatalf("%s did not close %s", closeKey, open)
			}
		}
	}
	m := press(newDemo(nil), "?", "j")
	if m.board.SelID != 1 {
		t.Fatal("keys must not reach the board while help is open")
	}
}

// TestTrueColorLayout checks that styled output keeps the exact screen
// geometry (ANSI-aware compositing) under a true-colour profile.
func TestTrueColorLayout(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii)
	for _, keys := range [][]string{nil, {"e"}, {"x"}, {"?"}, {"a"}} {
		m := press(newDemo(nil), keys...)
		requireScreen(t, m, DesignWidth, DesignHeight)
		if !strings.Contains(m.View(), "\x1b[") {
			t.Fatal("expected ANSI styling under TrueColor")
		}
	}
}
