// Package tui is the Bubbletea front end of task_prioritizer. The Model is the
// container: it owns the domain state, maps keys to domain operations and
// persists after every change. Rendering lives in the view*.go files.
package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/xilen0x/task_prioritizer_tui/internal/pomodoro"
	"github.com/xilen0x/task_prioritizer_tui/internal/tasks"
	"github.com/xilen0x/task_prioritizer_tui/internal/tui/theme"
)

// Design size of the terminal.
const (
	DesignWidth  = 132
	DesignHeight = 42
)

// saveEvery is how many ticks may pass between periodic saves of the
// time tracked by the pomodoro.
const saveEvery = 15

// SaveFunc persists the board and the pomodoro.
type SaveFunc func(tasks.Board, pomodoro.Timer) error

// AlertFunc tells the user that a pomodoro phase ended.
type AlertFunc func(title, body string)

// Options configures a Model.
type Options struct {
	Board tasks.Board
	Pomo  pomodoro.Timer
	// Save is called after every mutation and on quit; nil disables saving.
	Save SaveFunc
	// Clock returns the current time; nil means time.Now.
	Clock       func() time.Time
	Bar         theme.BarStyle
	PlainDigits bool
	// NoTick disables the one-second ticker (used by tests).
	NoTick bool
	// Alert is called when a work block or a break ends; nil disables alerts.
	Alert AlertFunc
}

// tickMsg is delivered once per second.
type tickMsg struct{}

// Model is the root Bubbletea model.
type Model struct {
	board       tasks.Board
	pomo        pomodoro.Timer
	save        SaveFunc
	alert       AlertFunc
	clock       func() time.Time
	now         time.Time
	bar         theme.BarStyle
	plainDigits bool
	noTick      bool

	width, height int

	editor    *editor // open edit modal, or nil
	confirmID int     // task awaiting delete confirmation, or 0
	help      bool
	about     bool

	ticksSinceSave int
	saveErr        error
}

// New builds a model from options.
func New(o Options) Model {
	clock := o.Clock
	if clock == nil {
		clock = time.Now
	}
	if o.Bar == "" {
		o.Bar = theme.Blocks
	}
	if o.Board.NextID == 0 && len(o.Board.Tasks) == 0 {
		o.Board = tasks.New()
	}
	if o.Pomo.Phase == "" {
		o.Pomo = pomodoro.New()
	}
	return Model{
		board: o.Board, pomo: o.Pomo, save: o.Save, alert: o.Alert, clock: clock, now: clock(),
		bar: o.Bar, plainDigits: o.PlainDigits, noTick: o.NoTick,
		width: DesignWidth, height: DesignHeight,
	}
}

// Board returns the current board (for tests and callers).
func (m Model) Board() tasks.Board { return m.board }

// Pomo returns the current pomodoro state.
func (m Model) Pomo() pomodoro.Timer { return m.pomo }

func (m Model) tick() tea.Cmd {
	if m.noTick {
		return nil
	}
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return tickMsg{} })
}

// Init starts the ticker.
func (m Model) Init() tea.Cmd { return m.tick() }

// persist saves the current state and records any error for display.
func (m *Model) persist() {
	m.ticksSinceSave = 0
	if m.save == nil {
		return
	}
	m.saveErr = m.save(m.board, m.pomo)
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tickMsg:
		ended := m.onTick()
		return m, tea.Batch(m.tick(), m.alertCmd(ended))
	case tea.KeyMsg:
		return m.onKey(msg)
	case tea.MouseMsg:
		return m.onMouse(msg)
	}
	return m, nil
}

// onTick advances the clock and, unless the editor is open, the pomodoro
// and the focused task's tracked time. It returns the phase that just
// ended, or "" when the phase did not change.
func (m *Model) onTick() (ended pomodoro.Phase) {
	m.now = m.clock()
	if m.editor != nil || !m.pomo.Running {
		return ""
	}
	prev := m.pomo.Phase
	var worked bool
	m.pomo, worked = m.pomo.Tick()
	if worked {
		m.board.TrackElapsed(1, m.now.Format(tasks.DateLayout))
	}
	m.ticksSinceSave++
	if m.pomo.Phase != prev || m.ticksSinceSave >= saveEvery {
		m.persist()
	}
	if m.pomo.Phase != prev {
		return prev
	}
	return ""
}

// alertCmd runs the alert for the phase that just ended, off the update loop.
func (m Model) alertCmd(ended pomodoro.Phase) tea.Cmd {
	if m.alert == nil || ended == "" {
		return nil
	}
	title, body := "Break over", fmt.Sprintf("Back to work: %d min.", pomodoro.WorkSeconds/60)
	if ended == pomodoro.Work {
		title, body = "Pomodoro done", fmt.Sprintf("Take a %d min break.", m.pomo.Left/60)
	}
	alert := m.alert
	return func() tea.Msg {
		alert(title, body)
		return nil
	}
}

func (m Model) quit() (tea.Model, tea.Cmd) {
	m.persist()
	return m, tea.Quit
}

// mutate applies fn to the model and saves.
func (m Model) mutate(fn func(*Model)) (tea.Model, tea.Cmd) {
	fn(&m)
	m.persist()
	return m, nil
}

func (m Model) onKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := k.String()
	if key == "ctrl+c" || key == "f10" {
		return m.quit()
	}
	switch {
	case m.confirmID != 0:
		return m.onConfirmKey(key)
	case m.editor != nil:
		return m.onEditorKey(k)
	case m.help || m.about:
		switch key {
		case "esc", "f1", "?", "a", "f9":
			m.help, m.about = false, false
		}
		return m, nil
	}
	return m.onMainKey(key)
}

func (m Model) onMainKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "down", "j":
		return m.mutate(func(m *Model) { m.board.Move(1) })
	case "up", "k":
		return m.mutate(func(m *Model) { m.board.Move(-1) })
	case "enter", "f5":
		return m.mutate(func(m *Model) { m.board.SetFocus() })
	case "e", "f3":
		m.openEdit()
	case "n", "f2":
		m.openNew()
	case "x", "delete", "f8":
		m.askDelete()
	case " ", "f4":
		return m.mutate(func(m *Model) { m.board.Cycle() })
	case "+", "=":
		return m.mutate(func(m *Model) { m.board.Bump(tasks.ProgressStep) })
	case "-":
		return m.mutate(func(m *Model) { m.board.Bump(-tasks.ProgressStep) })
	case "f1", "?":
		m.help = true
	case "a", "f9":
		m.about = true
	case "p", "f6":
		return m.mutate(func(m *Model) { m.pomo = m.pomo.Toggled() })
	case "r":
		return m.mutate(func(m *Model) { m.pomo = m.pomo.Reset() })
	case "q":
		// Alternative to F10, which some terminals (GNOME Terminal) capture.
		return m.quit()
	}
	return m, nil
}
