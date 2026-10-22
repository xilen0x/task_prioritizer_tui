package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/xilen0x/task_prioritizer_tui/internal/pomodoro"
	"github.com/xilen0x/task_prioritizer_tui/internal/tasks"
	"github.com/xilen0x/task_prioritizer_tui/internal/tui/theme"
)

// UI copy shared by several views.
var (
	statusLabel = map[tasks.Status]string{tasks.Todo: "TODO", tasks.Doing: "IN PROGRESS", tasks.Done: "DONE"}
	statusColor = map[tasks.Status]string{tasks.Todo: theme.OnDark3, tasks.Doing: theme.Doing, tasks.Done: theme.Success}
	prioLabel   = map[int]string{1: "P1 high", 2: "P2 medium", 3: "P3 low"}
	days        = []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	months      = []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
)

// Layout constants for the design size.
const (
	marginX    = 1
	sideWidth  = 30
	colGap     = 2
	focusBarW  = 62
	rowBarW    = 18
	chromeRows = 5 // header, gap, gap, hint, fkeys
)

func pctText(p int) string { return strconv.Itoa(p) + "%" }

// View renders the whole screen.
func (m Model) View() string {
	lines := m.screen()
	lines = m.overlay(lines)
	return strings.Join(lines, "\n")
}

// screen renders the main screen as exactly m.height lines of m.width.
func (m Model) screen() []string {
	l := m.layout()
	if l.tooSmall {
		return m.viewTooSmall()
	}
	if l.compact {
		m.plainDigits = true // the big font needs the full design size
	}
	cw, mainH := l.cw, l.mainH

	var out []string
	out = append(out, m.viewHeader(cw))
	out = append(out, blank(cw, theme.Ink1))
	main := hjoin(m.viewLeft(l), column(mainH, colGap), m.viewSide(l))
	out = append(out, main...)
	out = append(out, blank(cw, theme.Ink1))
	out = append(out, m.viewHint(cw), m.viewFKeys(cw, l.compact))
	for i := range out {
		out[i] = blank(marginX, theme.Ink1) + fit(out[i], cw, theme.Ink1) + blank(marginX, theme.Ink1)
	}
	return out
}

func (m Model) viewHeader(w int) string {
	bg := theme.Ink1
	left := st(theme.White, theme.Accent).Bold(true).Render(" TASKS ") +
		blank(2, bg) + paint(fmt.Sprintf("%d/%d tasks", len(m.board.Tasks), tasks.MaxTasks), theme.OnDark2, bg) +
		paint("  ·  ", theme.OnDark4, bg) + paint("max 4 on screen", theme.OnDark3, bg)
	n := m.now
	right := paint(fmt.Sprintf("%s %d %s", days[n.Weekday()], n.Day(), months[n.Month()-1]), theme.OnDark3, bg) +
		blank(2, bg) + bold(n.Format("15:04:05"), theme.OnDark1, bg)
	return fit(left, w-lenW(right), bg) + right
}

func (m Model) viewLeft(l layout) []string {
	w, h := l.leftW, l.mainH
	out := m.viewFocus(w, l.focusH)
	if len(m.board.Tasks) > 0 {
		others := m.others()
		for i := 0; i < 3; i++ {
			if l.rowGap > 0 {
				out = append(out, blank(w, theme.Ink1))
			}
			if i < len(others) {
				out = append(out, m.viewRow(others[i], w, l.rowH)...)
			} else {
				num := len(m.board.Tasks) + 1 + (i - len(others))
				out = append(out, viewEmptyRow(num, w, l.rowH)...)
			}
		}
	}
	for len(out) < h {
		out = append(out, blank(w, theme.Ink1))
	}
	return out[:h]
}

// others are the non-focused tasks in display order.
func (m Model) others() []tasks.Task {
	var out []tasks.Task
	for _, t := range m.board.Order() {
		if t.ID != m.board.FocusID {
			out = append(out, t)
		}
	}
	return out
}

// digits renders s with the big font, or as one plain bold line.
func (m Model) digits(s, fg string) []string {
	if m.plainDigits {
		return []string{bold(s, fg, theme.Ink1)}
	}
	var out []string
	for _, r := range theme.Big(s) {
		out = append(out, paint(r, fg, theme.Ink1))
	}
	return out
}

func (m Model) viewFocus(w, focusH int) []string {
	bg := theme.Ink1
	f, ok := m.board.Focused()
	num := "—"
	border := theme.OnDarkBorder
	if ok {
		num = strconv.Itoa(m.board.Num(f.ID))
		if f.ID == m.board.SelID {
			border = theme.Accent
		}
	}
	spec := boxSpec{width: w, title: "▶ FOCUS · TASK " + num, titleFg: theme.Accent, border: border, bg: bg, padX: 2}
	inner := w - 2 - 2*spec.padX
	body := focusH - 2
	if !ok {
		lines := make([]string, body)
		for i := range lines {
			lines[i] = blank(inner, bg)
		}
		lines[body/2-1] = center(bold("No tasks", theme.OnDark1, bg), inner, bg)
		lines[body/2+1] = center(paint("Press ", theme.OnDark3, bg)+bold("F2", theme.Accent, bg)+
			paint(" or ", theme.OnDark3, bg)+bold("n", theme.Accent, bg)+
			paint(" to create the first one.", theme.OnDark3, bg), inner, bg)
		return box(spec, lines)
	}

	color := string(theme.PctColor(f.Progress))
	var lines []string
	lines = append(lines, st(theme.Ink1, color).Bold(true).Render(" "+statusLabel[f.Status]+" ")+
		blank(2, bg)+paint(prioLabel[f.Prio], theme.OnDark2, bg)+paint("  ·  ", theme.OnDark4, bg)+
		paint("estimate "+theme.HoursLabel(f.Est), theme.OnDark3, bg))
	lines = append(lines, blank(inner, bg))
	title := wrapLines(f.Title, inner, 2)
	for i := 0; i < 2; i++ {
		s := ""
		if i < len(title) {
			s = bold(title[i], theme.OnDark1, bg)
		}
		lines = append(lines, s)
	}
	lines = append(lines, blank(inner, bg))
	lines = append(lines, m.viewCounters(f, color, inner)...)
	lines = append(lines, blank(inner, bg))
	fill, empty := theme.Bar(m.bar, f.Progress, min(focusBarW, inner-18))
	lines = append(lines, paint(fill, color, bg)+paint(empty, theme.Ink2, bg)+blank(2, bg)+paint("+/− adjusts 5%", theme.OnDark3, bg))
	for len(lines) < body {
		lines = append(lines, "")
	}
	return box(spec, lines[:body])
}

// viewCounters is the PROGRESS / POMODORO pair inside the focus box.
func (m Model) viewCounters(f tasks.Task, color string, w int) []string {
	bg := theme.Ink1
	p := m.pomo
	pomoFg, phase, run := theme.OnDark1, "WORK", "running"
	if p.Phase == pomodoro.Break {
		pomoFg, phase = theme.Success, "BREAK"
	}
	if !p.Running {
		run = "paused"
	}
	cyc := p.CycleDone()

	right := []string{paint("POMODORO · ", theme.OnDark3, bg) + bold(phase, pomoFg, bg) + paint(" · "+run, theme.OnDark3, bg)}
	right = append(right, m.digits(p.MMSS(), pomoFg)...)
	right = append(right, paint(strings.Repeat("●", cyc)+strings.Repeat("○", 4-cyc), pomoFg, bg)+blank(2, bg)+
		paint("p", theme.OnDark1, bg)+paint(" pause · ", theme.OnDark3, bg)+paint("r", theme.OnDark1, bg)+paint(" reset", theme.OnDark3, bg))

	left := []string{paint("PROGRESS", theme.OnDark3, bg)}
	left = append(left, m.digits(pctText(f.Progress), color)...)
	for len(left) < len(right) { // bottom-align like the design
		left = append([]string{""}, left...)
	}

	sep := paint("│", theme.OnDarkBorder, bg) + blank(3, bg)
	today := f.TodaySeconds(m.now.Format(tasks.DateLayout))
	clock := m.todayClock(today)
	leftW := maxW(left)
	if leftW+2*lenW(sep)+maxW(right)+3+maxW(clock) <= w {
		// Side by side: PROGRESS │ POMODORO │ TODAY.
		pw := maxW(right)
		for i := range right {
			right[i] = fit(right[i], pw, bg) + blank(3, bg) + sep + clock[i]
		}
	} else if fallback := paint("TODAY ", theme.OnDark3, bg) + bold(theme.HMS(today), theme.OnDark1, bg); left[0] == "" &&
		lenW(fallback) <= w-maxW(right)-lenW(sep) {
		// Too narrow for the TODAY column: use the free top line of PROGRESS.
		left[0] = fallback
	}

	rw := maxW(right)
	rightW := rw + lenW(sep)
	out := make([]string, len(right))
	for i := range right {
		out[i] = fit(left[i], w-rightW, bg) + sep + fit(right[i], rw, bg)
	}
	return out
}

// todayClock is the TODAY column: the focused task's time tracked today as
// a big h:mm with the exact hh:mm:ss below, as tall as the pomodoro column.
func (m Model) todayClock(seconds int) []string {
	bg := theme.Ink1
	out := []string{paint("TODAY", theme.OnDark3, bg)}
	out = append(out, m.digits(theme.TotalLabel(seconds), theme.Accent)...)
	return append(out, paint(theme.HMS(seconds)+" worked", theme.OnDark3, bg))
}

// maxW is the widest display width among lines.
func maxW(lines []string) int {
	w := 0
	for _, l := range lines {
		w = max(w, lenW(l))
	}
	return w
}

func (m Model) viewRow(t tasks.Task, w, h int) []string {
	selected := t.ID == m.board.SelID
	bg, border, numFg := theme.Ink1, theme.OnDarkRule, theme.OnDark3
	if selected {
		bg, border, numFg = theme.Ink2, theme.Accent, theme.Accent
	}
	spec := boxSpec{width: w, border: border, bg: bg, padX: 2}
	inner := w - 2 - 2*spec.padX
	color := string(theme.PctColor(t.Progress))
	numW, rightW := 4, rowBarW+2+5
	titleW := inner - numW - rightW - 1

	titleStyle := st(theme.OnDark1, bg).Bold(true)
	if t.Status == tasks.Done {
		titleStyle = titleStyle.Strikethrough(true)
	}
	fill, empty := theme.Bar(m.bar, t.Progress, rowBarW)
	lineA := fit(bold(strconv.Itoa(m.board.Num(t.ID)), numFg, bg), numW, bg) +
		fit(titleStyle.Render(ellipsis(t.Title, titleW)), titleW, bg) + blank(1, bg) +
		paint(fill, color, bg) + paint(empty, theme.Ink2, bg) + blank(2, bg) +
		fitRight(bold(pctText(t.Progress), color, bg), 5, bg)
	lineB := blank(numW, bg) + bold(statusLabel[t.Status], color, bg) +
		paint(" · ", theme.OnDark4, bg) + paint(prioLabel[t.Prio], theme.OnDark3, bg) +
		paint(" · ", theme.OnDark4, bg) + paint(theme.HMS(t.Elapsed)+" of "+theme.HoursLabel(t.Est), theme.OnDark3, bg)
	return box(spec, rowBody([]string{lineA, lineB}, h-2, inner, bg))
}

func viewEmptyRow(num, w, h int) []string {
	bg := theme.Ink1
	spec := boxSpec{width: w, border: theme.OnDarkRule, dashed: true, bg: bg, padX: 2}
	inner := w - 2 - 2*spec.padX
	line := fit(bold(strconv.Itoa(num), theme.Ink2, bg), 4, bg) + paint("free slot · ", theme.OnDark4, bg) +
		paint("F2", theme.Accent, bg) + paint(" new task", theme.OnDark4, bg)
	return box(spec, rowBody([]string{line, ""}, h-2, inner, bg))
}

// rowBody centres content vertically in n lines.
func rowBody(content []string, n, w int, bg string) []string {
	out := make([]string, n)
	top := max(0, (n-len(content))/2)
	for i := range out {
		out[i] = blank(w, bg)
		if j := i - top; j >= 0 && j < len(content) {
			out[i] = fit(content[j], w, bg)
		}
	}
	return out
}

func (m Model) viewSide(l layout) []string {
	w, h, barH := sideWidth, l.mainH, l.barH
	bg := theme.Ink1
	spec := boxSpec{width: w, title: "PROGRESS", titleFg: theme.OnDark2, border: theme.OnDarkBorder, bg: bg}
	inner := w - 2
	cellW := inner / 4
	body := []string{blank(inner, bg)}
	bars := make([][]string, 4)
	nums := ""
	pcts := ""
	for i := 0; i < 4; i++ {
		var t tasks.Task
		has := i < len(m.board.Tasks)
		color, numFg, pct := theme.Ink2, theme.Ink2, "—"
		if has {
			t = m.board.Tasks[i]
			color, numFg, pct = string(theme.PctColor(t.Progress)), theme.OnDark1, pctText(t.Progress)
			if t.ID == m.board.SelID {
				numFg = theme.Accent
			}
		}
		fill, empty := theme.Column(m.bar, t.Progress, barH)
		for _, e := range empty {
			bars[i] = append(bars[i], center(paint(e, theme.Ink2, bg), cellW, bg))
		}
		for _, f := range fill {
			bars[i] = append(bars[i], center(paint(f, color, bg), cellW, bg))
		}
		nums += center(bold(strconv.Itoa(i+1), numFg, bg), cellW, bg)
		pcts += center(paint(pct, theme.OnDark3, bg), cellW, bg)
	}
	for r := 0; r < barH; r++ {
		body = append(body, bars[0][r]+bars[1][r]+bars[2][r]+bars[3][r])
	}
	body = append(body, blank(inner, bg), nums, pcts)
	out := box(spec, body)

	out = append(out, blank(w, bg))
	s := m.board.Stats()
	items := []struct{ label, value, color string }{
		{"IN PROGRESS", strconv.Itoa(s.Doing), theme.Doing},
		{"TODO", strconv.Itoa(s.Todo), theme.OnDark2},
		{"DONE", strconv.Itoa(s.Done), theme.Success},
		{"TOTAL TIME", theme.TotalLabel(s.Elapsed), theme.OnDark1},
	}
	sumSpec := boxSpec{width: w, title: "SUMMARY", titleFg: theme.OnDark2, border: theme.OnDarkBorder, bg: bg, padX: 2}
	sumInner := w - 2 - 2*sumSpec.padX
	var rows []string
	for _, it := range items {
		v := bold(it.value, it.color, bg)
		rows = append(rows, fit(paint(it.label, theme.OnDark3, bg), sumInner-lenW(v), bg)+v)
	}
	sumH := h - len(out)
	out = append(out, box(sumSpec, spaceAround(rows, max(len(rows), sumH-2), sumInner, bg))...)
	return out[:min(len(out), h)]
}

func (m Model) viewHint(w int) string {
	bg := theme.Ink1
	if m.saveErr != nil {
		return fit(bold("Could not save: "+m.saveErr.Error(), theme.Accent, bg), w, bg)
	}
	hints := [][2]string{{"↑↓", "move"}, {"Enter", "focus"}, {"e", "edit"}, {"space", "change status"},
		{"+ −", "progress"}, {"x", "delete"}, {"p", "pomodoro"}, {"a", "about"}, {"q", "quit"}}
	line := ""
	for i, h := range hints {
		part := paint(h[0], theme.OnDark1, bg) + paint(" "+h[1], theme.OnDark3, bg)
		if i > 0 {
			part = blank(3, bg) + part
		}
		if lenW(line)+lenW(part) > w { // drop whole hints that do not fit
			break
		}
		line += part
	}
	return fit(line, w, bg)
}

// fkey is one entry of the function-key bar.
type fkey struct {
	key, label string
	disabled   bool
}

func (m Model) fkeys() []fkey {
	empty := len(m.board.Tasks) == 0
	return []fkey{
		{"F1", "Help", false}, {"F2", "New", m.board.Full()}, {"F3", "Edit", empty},
		{"F4", "Status", empty}, {"F5", "Focus", empty}, {"F8", "Delete", empty},
		{"F6", "Pomodoro", false}, {"F9", "About", false}, {"F10", "Quit", false},
	}
}

// disabledBg is Ink2 at 50% over Ink1 (the design dims disabled keys).
const disabledBg = "#293848"

// fkeySegments renders each function key; compact drops the padding so
// the bar fits a 100-column terminal. The widths drive mouse hit-testing.
func (m Model) fkeySegments(compact bool) []string {
	bg := theme.Ink1
	var out []string
	for _, f := range m.fkeys() {
		keyFg, labFg, labBg := theme.OnDark1, theme.OnDark2, theme.Ink2
		if f.disabled {
			keyFg, labFg, labBg = theme.OnDark4, theme.OnDark4, disabledBg
		}
		if compact {
			out = append(out, bold(f.key, keyFg, bg)+paint(" "+f.label+" ", labFg, labBg)+blank(1, bg))
		} else {
			out = append(out, bold(" "+f.key, keyFg, bg)+blank(1, bg)+paint(" "+f.label+"   ", labFg, labBg))
		}
	}
	return out
}

func (m Model) viewFKeys(w int, compact bool) string {
	return fit(strings.Join(m.fkeySegments(compact), ""), w, theme.Ink1)
}

// lenW is the display width of a styled string.
func lenW(s string) int { return lipgloss.Width(s) }
