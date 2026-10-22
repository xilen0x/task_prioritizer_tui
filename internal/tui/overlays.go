package tui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/xilen0x/task_prioritizer_tui/internal/tasks"
	"github.com/xilen0x/task_prioritizer_tui/internal/tui/theme"
)

// Editor fields in Tab order.
const (
	fieldTitle = iota
	fieldStatus
	fieldPrio
	fieldProgress
	fieldEst
	fieldCount
)

// Overlay widths, scaled from the design's pixel widths (1240px ≈ 132 cols).
const (
	modalW    = 74
	confirmW  = 60
	infoW     = 56
	labelColW = 16
	modalBarW = 26
)

// Scrim colours used to dim the screen behind an overlay.
const (
	scrimBg = "#191D24"
	scrimFg = "#3D424A"
)

// editor is the state of the edit/new modal.
type editor struct {
	isNew bool
	draft tasks.Task
	field int
	input textinput.Model
}

func newEditor(isNew bool, draft tasks.Task) *editor {
	in := textinput.New()
	in.Prompt = ""
	in.Placeholder = "Type a title"
	in.CharLimit = 120
	in.Width = modalW - 4 - labelColW - 3
	in.Cursor.SetMode(cursor.CursorStatic)
	in.SetValue(draft.Title)
	in.Focus()
	return &editor{isNew: isNew, draft: draft, field: fieldTitle, input: in}
}

func (e *editor) setField(f int) {
	e.field = f
	if f == fieldTitle {
		e.input.Focus()
	} else {
		e.input.Blur()
	}
}

// change steps the value of the active field by d, using the editor rules
// of the domain.
func (e *editor) change(d int) {
	switch e.field {
	case fieldStatus:
		e.draft = e.draft.StatusStepped(d)
	case fieldPrio:
		e.draft = e.draft.PrioStepped(d)
	case fieldProgress:
		e.draft = e.draft.ProgressStepped(d)
	case fieldEst:
		e.draft = e.draft.EstStepped(d)
	}
}

func (m *Model) openEdit() {
	if t, ok := m.board.Selected(); ok {
		m.editor = newEditor(false, t)
	}
}

func (m *Model) openNew() {
	if m.board.CanAdd() {
		m.editor = newEditor(true, tasks.NewDraft())
	}
}

func (m *Model) askDelete() {
	if _, ok := m.board.Selected(); ok {
		m.confirmID = m.board.SelID
	}
}

func (m Model) onConfirmKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "y", "Y", "enter":
		return m.mutate(func(m *Model) {
			m.board.Delete(m.confirmID)
			m.confirmID = 0
		})
	case "n", "N", "esc":
		m.confirmID = 0
	}
	return m, nil
}

func (m Model) onEditorKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	e := *m.editor // copy: the model is a value
	m.editor = &e
	switch k.String() {
	case "esc":
		m.editor = nil
		return m, nil
	case "enter":
		return m.mutate(func(m *Model) { m.saveEditor() })
	case "tab":
		e.setField((e.field + 1) % fieldCount)
		return m, nil
	case "shift+tab":
		e.setField((e.field + fieldCount - 1) % fieldCount)
		return m, nil
	case "down":
		e.setField(min(fieldCount-1, e.field+1))
		return m, nil
	case "up":
		e.setField(max(0, e.field-1))
		return m, nil
	case "left", "right":
		if e.field != fieldTitle {
			d := 1
			if k.String() == "left" {
				d = -1
			}
			e.change(d)
			return m, nil
		}
	}
	if e.field == fieldTitle {
		var cmd tea.Cmd
		e.input, cmd = e.input.Update(k)
		return m, cmd
	}
	return m, nil
}

// saveEditor stores the draft; empty titles become "Untitled task".
func (m *Model) saveEditor() {
	e := m.editor
	d := e.draft
	d.Title = e.input.Value()
	if e.isNew {
		m.board.Add(d)
	} else {
		m.board.Update(d)
	}
	m.editor = nil
}

// overlay composites the active dialog over a dimmed copy of base.
func (m Model) overlay(base []string) []string {
	var dlg []string
	switch {
	case m.confirmID != 0:
		dlg = m.viewConfirm()
	case m.editor != nil:
		dlg = m.viewEditor()
	case m.help:
		dlg = viewHelp()
	case m.about:
		dlg = m.viewAbout()
	default:
		return base
	}
	return composite(base, dlg)
}

// composite dims base like a scrim and centres dlg over it.
func composite(base, dlg []string) []string {
	out := make([]string, len(base))
	w := 0
	if len(base) > 0 {
		w = lenW(base[0])
	}
	dw := 0
	if len(dlg) > 0 {
		dw = lenW(dlg[0])
	}
	top, left := max(0, (len(base)-len(dlg))/2), max(0, (w-dw)/2)
	for i, line := range base {
		plain := ansi.Strip(line)
		j := i - top
		if j < 0 || j >= len(dlg) {
			out[i] = paint(plain, scrimFg, scrimBg)
			continue
		}
		l := ansi.Truncate(plain, left, "")
		r := ansi.TruncateLeft(plain, left+dw, "")
		out[i] = paint(l, scrimFg, scrimBg) + dlg[j] + paint(r, scrimFg, scrimBg)
	}
	return out
}

func (m Model) viewEditor() []string {
	e := m.editor
	bg := theme.Ink1
	title := "NEW TASK"
	if !e.isNew {
		title = "EDIT · TASK " + strconv.Itoa(m.board.Num(e.draft.ID))
	}
	spec := boxSpec{width: modalW, title: title, titleFg: theme.Accent, border: theme.Accent, bg: bg, padX: 2}
	inner := modalW - 2 - 2*spec.padX
	valW := inner - labelColW - 2
	d := e.draft

	row := func(f int, label string, value func(rowBg string) string) []string {
		rowBg, labFg, mark := bg, theme.OnDark3, " "
		if e.field == f {
			rowBg, labFg, mark = theme.Ink2, theme.Accent, "›"
		}
		content := blank(1, rowBg) + fit(paint(mark+" "+label, labFg, rowBg), labelColW, rowBg) + blank(1, rowBg) + fit(value(rowBg), valW, rowBg)
		return []string{blank(inner, rowBg), fit(content, inner, rowBg), blank(inner, rowBg)}
	}
	opt := func(label string, on bool, onBg, rowBg string) string {
		if on {
			return bold(" "+label+" ", theme.Ink1, onBg)
		}
		return bold(" "+label+" ", theme.OnDark3, rowBg)
	}

	var body []string
	body = append(body, blank(inner, bg))
	body = append(body, row(fieldTitle, "TITLE", func(rowBg string) string {
		in := e.input
		in.TextStyle = st(theme.OnDark1, rowBg).Bold(true)
		in.PlaceholderStyle = st(theme.OnDark4, rowBg)
		in.Cursor.Style = st(theme.Accent, rowBg)
		in.Cursor.TextStyle = st(theme.OnDark1, rowBg).Bold(true)
		return in.View()
	})...)
	body = append(body, row(fieldStatus, "STATUS", func(rowBg string) string {
		var parts []string
		for _, s := range tasks.Statuses() {
			parts = append(parts, opt(statusLabel[s], d.Status == s, statusColor[s], rowBg))
		}
		return strings.Join(parts, blank(1, rowBg))
	})...)
	body = append(body, row(fieldPrio, "PRIORITY", func(rowBg string) string {
		var parts []string
		for p := tasks.PrioMin; p <= tasks.PrioMax; p++ {
			parts = append(parts, opt(prioLabel[p], d.Prio == p, theme.OnDark1, rowBg))
		}
		return strings.Join(parts, blank(1, rowBg))
	})...)
	body = append(body, row(fieldProgress, "PROGRESS", func(rowBg string) string {
		color := string(theme.PctColor(d.Progress))
		fill, empty := theme.Bar(m.bar, d.Progress, modalBarW)
		return paint("◂", theme.Accent, rowBg) + blank(1, rowBg) + paint(fill, color, rowBg) + paint(empty, theme.Ink2, rowBg) +
			blank(1, rowBg) + paint("▸", theme.Accent, rowBg) + blank(2, rowBg) + bold(pctText(d.Progress), color, rowBg)
	})...)
	body = append(body, row(fieldEst, "ESTIMATE", func(rowBg string) string {
		return paint("◂", theme.Accent, rowBg) + blank(2, rowBg) + bold(theme.HoursLabel(d.Est), theme.OnDark1, rowBg) +
			blank(2, rowBg) + paint("▸", theme.Accent, rowBg)
	})...)
	body = append(body, blank(inner, bg), paint(strings.Repeat("─", inner), theme.OnDarkRule, bg))
	buttons := bold(" Enter Save ", theme.White, theme.Accent) + blank(1, bg) + paint(" Esc Cancel ", theme.OnDark1, theme.Ink2)
	body = append(body, fit(paint("Tab field · ←→ change value", theme.OnDark3, bg), inner-lenW(buttons), bg)+buttons)
	return box(spec, body)
}

func (m Model) viewConfirm() []string {
	bg := theme.Ink1
	spec := boxSpec{width: confirmW, title: "DELETE TASK", titleFg: theme.DangerBorder, border: theme.Danger, bg: bg, padX: 2}
	inner := confirmW - 2 - 2*spec.padX
	t := m.board.Get(m.confirmID)
	body := []string{""}
	for _, l := range wrapLines("Delete «"+t.Title+"»?", inner, 3) {
		body = append(body, bold(l, theme.OnDark1, bg))
	}
	body = append(body, "", paint("Progress and logged time will be lost.", theme.OnDark3, bg), "")
	buttons := bold(" Y Delete ", theme.White, theme.Danger) + blank(1, bg) + paint(" N Cancel ", theme.OnDark1, theme.Ink2)
	body = append(body, fitRight(buttons, inner, bg))
	return box(spec, body)
}

// infoRows renders a two-column grid with a fixed first column.
func infoRows(rows [][2]string, keyFg string, keyBold bool) []string {
	bg := theme.Ink1
	var out []string
	for _, r := range rows {
		k := paint(r[0], keyFg, bg)
		if keyBold {
			k = bold(r[0], keyFg, bg)
		}
		out = append(out, fit(k, labelColW, bg)+paint(r[1], theme.OnDark2, bg))
	}
	return out
}

func (m Model) viewAbout() []string {
	bg := theme.Ink1
	spec := boxSpec{width: infoW, title: "ABOUT", titleFg: theme.OnDark2, border: theme.OnDarkBorder, bg: bg, padX: 2}
	inner := infoW - 2 - 2*spec.padX
	done := m.pomo.Done
	doneLabel := strconv.Itoa(done) + " pomodoros completed"
	if done == 1 {
		doneLabel = "1 pomodoro completed"
	}
	body := []string{"", bold("task_prioritizer", theme.OnDark1, bg) + blank(2, bg) + bold("v0.3.0", theme.Accent, bg), ""}
	for _, l := range wrapLines("A task monitor for the terminal. Up to 4 tasks on screen and one in focus, with a pomodoro timer.", inner, 4) {
		body = append(body, paint(l, theme.OnDark2, bg))
	}
	body = append(body, "")
	body = append(body, infoRows([][2]string{
		{"Pomodoro", "25 min work · 5 min break"},
		{"Long break", "15 min every 4 pomodoros"},
		{"Today", doneLabel},
		{"Author", "Carlos Astorga"},
	}, theme.OnDark3, false)...)
	body = append(body, "", paint("Esc", theme.OnDark1, bg)+paint(" close", theme.OnDark3, bg))
	return box(spec, body)
}

func viewHelp() []string {
	bg := theme.Ink1
	spec := boxSpec{width: infoW, title: "HELP", titleFg: theme.OnDark2, border: theme.OnDarkBorder, bg: bg, padX: 2}
	inner := infoW - 2 - 2*spec.padX
	body := []string{""}
	body = append(body, infoRows([][2]string{
		{"↑ ↓  j k", "Move the selection"},
		{"Enter  F5", "Put the task in focus"},
		{"n  F2", "New task (max 4)"},
		{"e  F3", "Edit the selected task"},
		{"space  F4", "To do → in progress → done"},
		{"+  −", "Progress ±5%"},
		{"x  F8", "Delete (y confirms)"},
		{"p  F6", "Pause or resume the pomodoro"},
		{"r", "Reset the pomodoro"},
		{"a  F9", "About"},
		{"q  F10", "Quit"},
	}, theme.Accent, true)...)
	body = append(body, "")
	for _, l := range wrapLines("The focused task logs time only in work blocks.", inner, 2) {
		body = append(body, paint(l, theme.OnDark3, bg))
	}
	return box(spec, body)
}
