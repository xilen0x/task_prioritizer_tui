package tasks

// Board holds the tasks plus the focused and selected task ids. IDs are
// never reused; NextID is the id for the next created task. An id of 0
// means "none".
type Board struct {
	Tasks   []Task `json:"tasks"`
	FocusID int    `json:"focusId"`
	SelID   int    `json:"selId"`
	NextID  int    `json:"nextId"`
}

// New returns an empty board.
func New() Board { return Board{NextID: 1} }

// Demo returns the sample board shown in the design.
func Demo() Board {
	return Board{
		Tasks: []Task{
			{ID: 1, Title: "Migrate the database to PostgreSQL 16", Status: Doing, Prio: 1, Progress: 70, Est: 6, Elapsed: 5047},
			{ID: 2, Title: "Review the backups module PR", Status: Doing, Prio: 2, Progress: 40, Est: 2, Elapsed: 1820},
			{ID: 3, Title: "Update the server packages", Status: Todo, Prio: 3, Progress: 0, Est: 1, Elapsed: 0},
			{ID: 4, Title: "Write the v0.3 changelog", Status: Done, Prio: 3, Progress: 100, Est: 1, Elapsed: 2700},
		},
		FocusID: 1, SelID: 1, NextID: 5,
	}
}

func (b Board) index(id int) int {
	for i, t := range b.Tasks {
		if t.ID == id {
			return i
		}
	}
	return -1
}

// Get returns the task with id, or the zero Task when absent.
func (b Board) Get(id int) Task {
	if i := b.index(id); i >= 0 {
		return b.Tasks[i]
	}
	return Task{}
}

// Has reports whether a task with id exists.
func (b Board) Has(id int) bool { return b.index(id) >= 0 }

// Focused returns the task in focus.
func (b Board) Focused() (Task, bool) {
	i := b.index(b.FocusID)
	if i < 0 {
		return Task{}, false
	}
	return b.Tasks[i], true
}

// Selected returns the selected task.
func (b Board) Selected() (Task, bool) {
	i := b.index(b.SelID)
	if i < 0 {
		return Task{}, false
	}
	return b.Tasks[i], true
}

// Num is the 1-based slot number of a task (its creation position), or 0.
func (b Board) Num(id int) int { return b.index(id) + 1 }

// Full reports whether the board holds MaxTasks tasks.
func (b Board) Full() bool { return len(b.Tasks) >= MaxTasks }

// CanAdd reports whether another task fits.
func (b Board) CanAdd() bool { return !b.Full() }

// Order is the display order: the focused task first, then the rest in
// slot order.
func (b Board) Order() []Task {
	f, ok := b.Focused()
	if !ok {
		return append([]Task(nil), b.Tasks...)
	}
	out := []Task{f}
	for _, t := range b.Tasks {
		if t.ID != f.ID {
			out = append(out, t)
		}
	}
	return out
}

// Move shifts the selection by d positions in display order, wrapping.
func (b *Board) Move(d int) {
	o := b.Order()
	n := len(o)
	if n == 0 {
		return
	}
	i := 0
	for j, t := range o {
		if t.ID == b.SelID {
			i = j
			break
		}
	}
	b.SelID = o[((i+d)%n+n)%n].ID
}

// Select selects the task with id when it exists.
func (b *Board) Select(id int) {
	if b.Has(id) {
		b.SelID = id
	}
}

// SetFocus puts the selected task in focus.
func (b *Board) SetFocus() {
	if b.Has(b.SelID) {
		b.FocusID = b.SelID
	}
}

func (b *Board) patchSelected(fn func(Task) Task) {
	if i := b.index(b.SelID); i >= 0 {
		b.Tasks[i] = fn(b.Tasks[i])
	}
}

// Cycle advances the status of the selected task.
func (b *Board) Cycle() { b.patchSelected(Task.Cycled) }

// Bump changes the selected task's progress by delta percent.
func (b *Board) Bump(delta int) {
	b.patchSelected(func(t Task) Task { return t.Bumped(delta) })
}

// Add stores draft as a new task and selects it. The first task on an
// empty board also takes focus. It returns false when the board is full.
func (b *Board) Add(draft Task) bool {
	if b.Full() {
		return false
	}
	if b.NextID < 1 {
		b.NextID = 1
	}
	draft.ID = b.NextID
	draft.Title = normalizeTitle(draft.Title)
	if len(b.Tasks) == 0 {
		b.FocusID = draft.ID
	}
	b.Tasks = append(b.Tasks, draft)
	b.SelID = draft.ID
	b.NextID++
	return true
}

// Update replaces the task with the same id, keeping its slot.
func (b *Board) Update(t Task) {
	if i := b.index(t.ID); i >= 0 {
		t.Title = normalizeTitle(t.Title)
		b.Tasks[i] = t
	}
}

// Delete removes the task with id. If the focus disappears it moves to the
// first remaining task; the selection follows the focus.
func (b *Board) Delete(id int) {
	i := b.index(id)
	if i < 0 {
		return
	}
	b.Tasks = append(b.Tasks[:i:i], b.Tasks[i+1:]...)
	if !b.Has(b.FocusID) {
		b.FocusID = 0
		if len(b.Tasks) > 0 {
			b.FocusID = b.Tasks[0].ID
		}
	}
	b.SelID = b.FocusID
}

// TrackElapsed adds seconds to the focused task when it is in progress,
// both to its all-time Elapsed and to its daily counter for date (a local
// "2006-01-02" date). A new date restarts the daily counter. The caller
// decides whether the pomodoro is in a running work block.
func (b *Board) TrackElapsed(seconds int, date string) {
	i := b.index(b.FocusID)
	if i < 0 || b.Tasks[i].Status != Doing {
		return
	}
	t := &b.Tasks[i]
	t.Elapsed += seconds
	if t.TodayDate != date {
		t.Today, t.TodayDate = 0, date
	}
	t.Today += seconds
}

// Stats summarises the board.
type Stats struct {
	Doing, Todo, Done int
	Elapsed           int
}

// Stats counts tasks per status and totals tracked seconds.
func (b Board) Stats() Stats {
	var s Stats
	for _, t := range b.Tasks {
		switch t.Status {
		case Doing:
			s.Doing++
		case Todo:
			s.Todo++
		case Done:
			s.Done++
		}
		s.Elapsed += t.Elapsed
	}
	return s
}
