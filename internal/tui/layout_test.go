package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/golden"
)

func resize(m Model, w, h int) Model {
	next, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return next.(Model)
}

func TestResizeGoldens(t *testing.T) {
	tests := []struct {
		name string
		w, h int
	}{
		{"compact_100x30", 100, 30},
		{"compact_120x36", 120, 36},
		{"large_150x50", 150, 50},
		{"too_small_80x24", 80, 24},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := resize(newDemo(nil), tt.w, tt.h)
			requireScreen(t, m, tt.w, tt.h)
			golden.RequireEqual(t, []byte(m.View()))
		})
	}
}

func TestCompactUsesPlainDigits(t *testing.T) {
	m := resize(newDemo(nil), 120, 40)
	v := m.View()
	if !strings.Contains(v, "17:43") || strings.Contains(v, "██████ ██████") {
		t.Fatal("compact layout must render plain digits")
	}
}

func TestOverlayFitsCompactScreen(t *testing.T) {
	for _, k := range []string{"e", "x", "?", "a"} {
		m := press(resize(newDemo(nil), MinWidth, MinHeight), k)
		requireScreen(t, m, MinWidth, MinHeight)
	}
}

func TestTooSmallStillQuits(t *testing.T) {
	m := resize(newDemo(nil), 60, 20)
	if !strings.Contains(m.View(), "Terminal too small") {
		t.Fatal("missing too-small notice")
	}
	_, cmd := m.Update(key("f10"))
	if cmd == nil {
		t.Fatal("F10 must quit even when the terminal is too small")
	}
}

func click(m Model, x, y int) Model {
	next, _ := m.Update(tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	return next.(Model)
}

func TestMouse(t *testing.T) {
	l := newDemo(nil).layout()
	t.Run("click on a row selects it", func(t *testing.T) {
		m := click(newDemo(nil), 10, l.rowTop(1)+1) // second row: task 3
		if m.board.SelID != 3 {
			t.Fatalf("sel = %d", m.board.SelID)
		}
	})
	t.Run("click on focus box selects focus", func(t *testing.T) {
		m := click(press(newDemo(nil), "down"), 10, headerRows+3)
		if m.board.SelID != 1 {
			t.Fatalf("sel = %d", m.board.SelID)
		}
	})
	t.Run("click on sidebar column selects task", func(t *testing.T) {
		cellW := (sideWidth - 2) / 4
		m := click(newDemo(nil), l.sideX()+1+3*cellW+2, headerRows+5)
		if m.board.SelID != 4 {
			t.Fatalf("sel = %d", m.board.SelID)
		}
	})
	t.Run("click on free slot opens new task editor", func(t *testing.T) {
		m := press(newDemo(nil), "x", "y") // 3 tasks: one free slot
		m = click(m, 10, l.rowTop(2)+1)
		if m.editor == nil || !m.editor.isNew {
			t.Fatal("editor not opened")
		}
	})
	t.Run("click on F1 opens help and any click closes it", func(t *testing.T) {
		m := click(newDemo(nil), 2, l.h-1)
		if !m.help {
			t.Fatal("help not opened")
		}
		if m = click(m, 50, 10); m.help {
			t.Fatal("help not closed")
		}
	})
	t.Run("disabled F2 does nothing", func(t *testing.T) {
		segW := lenW(newDemo(nil).fkeySegments(false)[0])
		m := click(newDemo(nil), marginX+segW+1, l.h-1)
		if m.editor != nil {
			t.Fatal("F2 is disabled with 4 tasks")
		}
	})
	t.Run("wheel moves selection", func(t *testing.T) {
		next, _ := newDemo(nil).Update(tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonWheelDown})
		if next.(Model).board.SelID != 2 {
			t.Fatal("wheel down did not move")
		}
	})
	t.Run("clicks are ignored while editing", func(t *testing.T) {
		m := click(press(newDemo(nil), "e"), 10, l.rowTop(1)+1)
		if m.editor == nil || m.board.SelID != 1 {
			t.Fatal("click leaked through the editor")
		}
	})
}

func TestTodayClockShowsFocusedTaskTimeForToday(t *testing.T) {
	tests := []struct {
		name string
		w, h int
	}{
		{"compact", 100, 30},
		{"design", 132, 44},
		{"large", 150, 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newDemo(nil)
			m.board.Tasks[0].Today, m.board.Tasks[0].TodayDate = 3723, "2026-10-18"
			m = resize(m, tt.w, tt.h)
			requireScreen(t, m, tt.w, tt.h)
			v := m.View()
			if !strings.Contains(v, "TODAY") || !strings.Contains(v, "01:02:03") {
				t.Fatalf("TODAY clock with 01:02:03 missing:\n%s", v)
			}
		})
	}
}

func TestTodayClockIgnoresStaleDate(t *testing.T) {
	m := newDemo(nil)
	m.board.Tasks[0].Today, m.board.Tasks[0].TodayDate = 3723, "2026-10-17"
	m = resize(m, 150, 50)
	if v := m.View(); strings.Contains(v, "01:02:03") || !strings.Contains(v, "00:00:00") {
		t.Fatalf("stale daily counter must show as zero:\n%s", v)
	}
}

func TestTodayClockFallsBackWhenNarrow(t *testing.T) {
	m := newDemo(nil)
	m.board.Tasks[0].Today, m.board.Tasks[0].TodayDate = 3723, "2026-10-18"
	m = resize(m, 150, 50)
	f, _ := m.board.Focused()
	const w = 60 // too narrow for the big TODAY digits beside the pomodoro
	lines := m.viewCounters(f, "#ffffff", w)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "TODAY 01:02:03") {
		t.Fatalf("narrow fallback missing:\n%s", joined)
	}
	for _, l := range lines {
		if lenW(l) > w {
			t.Fatalf("line overflows %d: %q", w, l)
		}
	}
}
