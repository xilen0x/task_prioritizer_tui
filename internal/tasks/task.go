// Package tasks is the task-board domain: up to four tasks, one in focus,
// with the progress and status rules of the tasks design. It has no UI
// dependencies.
package tasks

import "strings"

// Status is the lifecycle state of a task.
type Status string

// Task statuses, in cycle order.
const (
	Todo  Status = "todo"
	Doing Status = "doing"
	Done  Status = "done"
)

var statusOrder = []Status{Todo, Doing, Done}

// Statuses returns the statuses in cycle order.
func Statuses() []Status { return append([]Status(nil), statusOrder...) }

func (s Status) index() int {
	for i, x := range statusOrder {
		if x == s {
			return i
		}
	}
	return 0
}

// Limits and defaults of the domain.
const (
	MaxTasks     = 4
	ProgressStep = 10
	EstStep      = 0.5
	EstMin       = 0.5
	EstMax       = 40
	PrioMin      = 1
	PrioMax      = 3
	UntitledName = "Untitled task"
)

// Task is one item on the board. Progress is 0..100, Est is in hours and
// Elapsed is tracked time in seconds.
type Task struct {
	ID       int     `json:"id"`
	Title    string  `json:"title"`
	Status   Status  `json:"status"`
	Prio     int     `json:"prio"`
	Progress int     `json:"progress"`
	Est      float64 `json:"est"`
	Elapsed  int     `json:"elapsed"`
}

// NewDraft returns the defaults for a task created from the editor.
func NewDraft() Task {
	return Task{Status: Todo, Prio: 2, Est: 1}
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// Cycled advances todo → doing → done → todo. Becoming done forces
// 100%; leaving done drops 100% to 90%.
func (t Task) Cycled() Task {
	next := statusOrder[(t.Status.index()+1)%len(statusOrder)]
	switch {
	case next == Done:
		t.Progress = 100
	case t.Progress == 100:
		t.Progress = 90
	}
	t.Status = next
	return t
}

// Bumped changes progress by delta (clamped to 0..100) and derives the
// status: 100% is done, any change on a done task returns it to doing and
// a todo task with progress becomes doing.
func (t Task) Bumped(delta int) Task {
	p := clamp(t.Progress+delta, 0, 100)
	switch {
	case p == 100:
		t.Status = Done
	case t.Status == Done:
		t.Status = Doing
	case t.Status == Todo && p > 0:
		t.Status = Doing
	}
	t.Progress = p
	return t
}

// StatusStepped moves the status by d positions (wrapping), as the editor
// does. Choosing done forces 100%.
func (t Task) StatusStepped(d int) Task {
	n := len(statusOrder)
	t.Status = statusOrder[((t.Status.index()+d)%n+n)%n]
	if t.Status == Done {
		t.Progress = 100
	}
	return t
}

// WithStatus sets the status directly, as clicking an editor option does.
func (t Task) WithStatus(s Status) Task {
	t.Status = s
	if s == Done {
		t.Progress = 100
	}
	return t
}

// PrioStepped moves the priority by d, clamped to 1..3.
func (t Task) PrioStepped(d int) Task {
	t.Prio = clamp(t.Prio+d, PrioMin, PrioMax)
	return t
}

// ProgressStepped moves progress by d steps of 10% as the editor does:
// 100% is done and lowering a done task returns it to doing.
func (t Task) ProgressStepped(d int) Task {
	p := clamp(t.Progress+d*ProgressStep, 0, 100)
	switch {
	case p == 100:
		t.Status = Done
	case t.Status == Done:
		t.Status = Doing
	}
	t.Progress = p
	return t
}

// EstStepped moves the estimate by d half hours, clamped to 0.5..40.
func (t Task) EstStepped(d int) Task {
	e := t.Est + float64(d)*EstStep
	if e < EstMin {
		e = EstMin
	}
	if e > EstMax {
		e = EstMax
	}
	t.Est = e
	return t
}

func normalizeTitle(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return UntitledName
	}
	return s
}
