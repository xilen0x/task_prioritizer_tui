// Package pomodoro is a pure pomodoro state machine: 25 minutes of work,
// then a 5 minute break, or a 15 minute break after every 4 pomodoros.
// It never reads the clock; the caller ticks it once per second.
package pomodoro

import "fmt"

// Phase is the current block type.
type Phase string

// Phases.
const (
	Work  Phase = "work"
	Break Phase = "break"
)

// Block lengths in seconds.
const (
	WorkSeconds       = 25 * 60
	ShortBreakSeconds = 5 * 60
	LongBreakSeconds  = 15 * 60
	LongBreakEvery    = 4
)

// Timer is the pomodoro state. Left is the remaining seconds of the
// current block and Done counts completed work blocks.
type Timer struct {
	Phase   Phase `json:"phase"`
	Left    int   `json:"left"`
	Running bool  `json:"running"`
	Done    int   `json:"done"`
}

// New returns a running timer at the start of a work block.
func New() Timer { return Timer{Phase: Work, Left: WorkSeconds, Running: true} }

// Tick advances the timer by one second. worked reports whether that
// second belonged to a running work block (and so counts as task time).
func (t Timer) Tick() (next Timer, worked bool) {
	if !t.Running {
		return t, false
	}
	worked = t.Phase == Work
	switch {
	case t.Left > 1:
		t.Left--
	case t.Phase == Work:
		t.Done++
		t.Phase = Break
		t.Left = ShortBreakSeconds
		if t.Done%LongBreakEvery == 0 {
			t.Left = LongBreakSeconds
		}
	default:
		t.Phase = Work
		t.Left = WorkSeconds
	}
	return t, worked
}

// Toggled pauses or resumes the timer.
func (t Timer) Toggled() Timer {
	t.Running = !t.Running
	return t
}

// Reset returns to a paused, full work block, keeping the completed count.
func (t Timer) Reset() Timer {
	t.Phase, t.Left, t.Running = Work, WorkSeconds, false
	return t
}

// MMSS formats the remaining time as mm:ss.
func (t Timer) MMSS() string { return fmt.Sprintf("%02d:%02d", t.Left/60, t.Left%60) }

// CycleDone is the number of completed pomodoros in the current set of 4.
func (t Timer) CycleDone() int { return t.Done % LongBreakEvery }
