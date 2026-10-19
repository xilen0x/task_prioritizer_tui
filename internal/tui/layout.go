package tui

import (
	"strconv"

	"github.com/xilenox/tareas/internal/tui/theme"
)

// Minimum usable terminal size; below it only a notice is shown.
const (
	MinWidth  = 100
	MinHeight = 30
)

// Screen geometry shared by the views and the mouse hit-testing.
const (
	headerRows = 2 // header line + gap; the main area starts below
	sideChrome = 6 // PROGRESS box: borders, top gap, number and % rows
	minSummary = 6 // SUMMARY box: borders + four items
	minRowH    = 4 // task row: borders + title and meta lines
)

// layout is the computed geometry for the current terminal size.
type layout struct {
	w, h     int
	tooSmall bool
	compact  bool // smaller than the design: plain digits, tighter rows
	cw       int  // content width (inside the side margins)
	leftW    int
	mainH    int
	focusH   int
	rowH     int
	rowGap   int
	barH     int // height of the sidebar's vertical bars
}

func (m Model) layout() layout {
	l := layout{w: m.width, h: m.height}
	l.tooSmall = l.w < MinWidth || l.h < MinHeight
	l.compact = l.w < DesignWidth || l.h < DesignHeight
	l.cw = l.w - 2*marginX
	l.leftW = l.cw - sideWidth - colGap
	l.mainH = l.h - chromeRows
	l.focusH = 16
	if l.compact || m.plainDigits {
		l.focusH = 12
	}
	avail := l.mainH - l.focusH
	l.rowGap = 1
	if avail < 3*(minRowH+1) {
		l.rowGap = 0
	}
	l.rowH = max(minRowH, avail/3-l.rowGap)
	l.barH = max(3, min(13, l.mainH-sideChrome-1-minSummary))
	return l
}

// rowTop is the screen line of task row i (0..2).
func (l layout) rowTop(i int) int {
	return headerRows + l.focusH + i*(l.rowGap+l.rowH) + l.rowGap
}

// sideX is the first screen column of the sidebar.
func (l layout) sideX() int { return marginX + l.leftW + colGap }

// viewTooSmall is shown when the terminal is below MinWidth×MinHeight.
func (m Model) viewTooSmall() []string {
	w, h := max(1, m.width), max(1, m.height)
	bg := theme.Ink1
	lines := column(h, w)
	msg := []string{
		bold("Terminal too small", theme.OnDark1, bg),
		paint("At least "+strconv.Itoa(MinWidth)+"×"+strconv.Itoa(MinHeight)+
			" is needed (current "+strconv.Itoa(m.width)+"×"+strconv.Itoa(m.height)+")", theme.OnDark3, bg),
		bold("F10", theme.Accent, bg) + paint(" quit", theme.OnDark3, bg),
	}
	top := max(0, (h-len(msg))/2)
	for i, s := range msg {
		if top+i < h {
			lines[top+i] = center(s, w, bg)
		}
	}
	return lines
}
