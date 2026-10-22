package tasks

import (
	"reflect"
	"testing"
)

func ids(ts []Task) []int {
	out := make([]int, len(ts))
	for i, t := range ts {
		out[i] = t.ID
	}
	return out
}

func TestOrderPutsFocusFirst(t *testing.T) {
	b := Demo()
	b.FocusID = 3
	if got, want := ids(b.Order()), []int{3, 1, 2, 4}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Order() = %v, want %v", got, want)
	}
}

func TestMoveWrapsInDisplayOrder(t *testing.T) {
	tests := []struct {
		name    string
		focus   int
		sel     int
		delta   int
		wantSel int
	}{
		{"down from focus", 1, 1, 1, 2},
		{"up from focus wraps to last", 1, 1, -1, 4},
		{"down from last wraps to focus", 1, 4, 1, 1},
		{"focus elsewhere changes order", 3, 3, 1, 1},
		{"unknown selection starts at first", 1, 99, 1, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Demo()
			b.FocusID, b.SelID = tt.focus, tt.sel
			b.Move(tt.delta)
			if b.SelID != tt.wantSel {
				t.Fatalf("SelID = %d, want %d", b.SelID, tt.wantSel)
			}
		})
	}
}

func TestMoveOnEmptyBoardIsNoop(t *testing.T) {
	b := New()
	b.Move(1)
	if b.SelID != 0 {
		t.Fatalf("SelID = %d, want 0", b.SelID)
	}
}

func TestCycle(t *testing.T) {
	tests := []struct {
		name         string
		in           Task
		wantStatus   Status
		wantProgress int
	}{
		{"todo to doing keeps progress", Task{Status: Todo, Progress: 20}, Doing, 20},
		{"doing to done forces 100", Task{Status: Doing, Progress: 40}, Done, 100},
		{"done to todo drops 100 to 90", Task{Status: Done, Progress: 100}, Todo, 90},
		{"done below 100 keeps progress", Task{Status: Done, Progress: 70}, Todo, 70},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in.Cycled()
			if got.Status != tt.wantStatus || got.Progress != tt.wantProgress {
				t.Fatalf("Cycled() = %s/%d, want %s/%d", got.Status, got.Progress, tt.wantStatus, tt.wantProgress)
			}
		})
	}
}

func TestBump(t *testing.T) {
	tests := []struct {
		name         string
		in           Task
		delta        int
		wantStatus   Status
		wantProgress int
	}{
		{"reaching 100 sets done", Task{Status: Doing, Progress: 90}, 10, Done, 100},
		{"clamps at 100", Task{Status: Done, Progress: 100}, 10, Done, 100},
		{"lowering done sets doing", Task{Status: Done, Progress: 100}, -10, Doing, 90},
		{"todo above 0 sets doing", Task{Status: Todo, Progress: 0}, 10, Doing, 10},
		{"todo at 0 stays todo", Task{Status: Todo, Progress: 0}, -10, Todo, 0},
		{"doing stays doing", Task{Status: Doing, Progress: 50}, -10, Doing, 40},
		{"doing down to 0 stays doing", Task{Status: Doing, Progress: 10}, -10, Doing, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in.Bumped(tt.delta)
			if got.Status != tt.wantStatus || got.Progress != tt.wantProgress {
				t.Fatalf("Bumped(%d) = %s/%d, want %s/%d", tt.delta, got.Status, got.Progress, tt.wantStatus, tt.wantProgress)
			}
		})
	}
}

func TestBoardCycleAndBumpActOnSelection(t *testing.T) {
	b := Demo()
	b.SelID = 3
	b.Cycle()
	if got := b.Get(3).Status; got != Doing {
		t.Fatalf("after Cycle status = %s, want doing", got)
	}
	b.Bump(10)
	if got := b.Get(3).Progress; got != 10 {
		t.Fatalf("after Bump progress = %d, want 10", got)
	}
	if b.Get(1).Progress != 70 {
		t.Fatal("unselected task must not change")
	}
}

func TestDraftSteps(t *testing.T) {
	t.Run("status step wraps and done forces 100", func(t *testing.T) {
		d := Task{Status: Todo, Progress: 30}
		d = d.StatusStepped(-1)
		if d.Status != Done || d.Progress != 100 {
			t.Fatalf("got %s/%d", d.Status, d.Progress)
		}
		d = d.StatusStepped(1)
		if d.Status != Todo || d.Progress != 100 {
			t.Fatalf("leaving done in editor keeps progress, got %s/%d", d.Status, d.Progress)
		}
	})
	t.Run("prio clamps 1..3", func(t *testing.T) {
		d := Task{Prio: 1}
		if d.PrioStepped(-1).Prio != 1 || d.PrioStepped(1).Prio != 2 || (Task{Prio: 3}).PrioStepped(1).Prio != 3 {
			t.Fatal("prio clamp failed")
		}
	})
	t.Run("progress step does not promote todo", func(t *testing.T) {
		d := Task{Status: Todo}.ProgressStepped(1)
		if d.Status != Todo || d.Progress != 10 {
			t.Fatalf("got %s/%d", d.Status, d.Progress)
		}
		d = Task{Status: Done, Progress: 100}.ProgressStepped(-1)
		if d.Status != Doing || d.Progress != 90 {
			t.Fatalf("got %s/%d", d.Status, d.Progress)
		}
		d = Task{Status: Doing, Progress: 90}.ProgressStepped(1)
		if d.Status != Done {
			t.Fatalf("got %s", d.Status)
		}
	})
	t.Run("estimate steps 0.5 within 0.5..40", func(t *testing.T) {
		if got := (Task{Est: 0.5}).EstStepped(-1).Est; got != 0.5 {
			t.Fatalf("lower clamp got %v", got)
		}
		if got := (Task{Est: 40}).EstStepped(1).Est; got != 40 {
			t.Fatalf("upper clamp got %v", got)
		}
		if got := (Task{Est: 1}).EstStepped(1).Est; got != 1.5 {
			t.Fatalf("step got %v", got)
		}
	})
}

func TestAdd(t *testing.T) {
	t.Run("first task becomes focus and selection", func(t *testing.T) {
		b := New()
		if !b.Add(NewDraft()) {
			t.Fatal("Add returned false")
		}
		got := b.Tasks[0]
		if got.ID != 1 || b.FocusID != 1 || b.SelID != 1 || b.NextID != 2 {
			t.Fatalf("got id=%d focus=%d sel=%d next=%d", got.ID, b.FocusID, b.SelID, b.NextID)
		}
		if got.Title != "Untitled task" {
			t.Fatalf("empty title stored as %q", got.Title)
		}
	})
	t.Run("later task keeps focus and gets selection", func(t *testing.T) {
		b := New()
		b.Add(Task{Title: "  uno  ", Prio: 2, Est: 1})
		b.Add(Task{Title: "dos", Prio: 2, Est: 1})
		if b.FocusID != 1 || b.SelID != 2 {
			t.Fatalf("focus=%d sel=%d", b.FocusID, b.SelID)
		}
		if b.Tasks[0].Title != "uno" {
			t.Fatalf("title not trimmed: %q", b.Tasks[0].Title)
		}
	})
	t.Run("refuses a fifth task", func(t *testing.T) {
		b := Demo()
		if b.CanAdd() || b.Add(NewDraft()) || len(b.Tasks) != MaxTasks {
			t.Fatal("board accepted more than MaxTasks")
		}
	})
}

func TestUpdate(t *testing.T) {
	b := Demo()
	d := b.Get(2)
	d.Title = "   "
	d.Prio = 1
	b.Update(d)
	got := b.Get(2)
	if got.Title != "Untitled task" || got.Prio != 1 {
		t.Fatalf("got %+v", got)
	}
	if ids(b.Tasks)[1] != 2 {
		t.Fatal("update must keep position")
	}
}

func TestDelete(t *testing.T) {
	t.Run("deleting focus moves focus to first remaining", func(t *testing.T) {
		b := Demo()
		b.Delete(1)
		if b.FocusID != 2 || b.SelID != 2 || len(b.Tasks) != 3 {
			t.Fatalf("focus=%d sel=%d len=%d", b.FocusID, b.SelID, len(b.Tasks))
		}
	})
	t.Run("deleting other keeps focus and selects it", func(t *testing.T) {
		b := Demo()
		b.SelID = 3
		b.Delete(3)
		if b.FocusID != 1 || b.SelID != 1 {
			t.Fatalf("focus=%d sel=%d", b.FocusID, b.SelID)
		}
	})
	t.Run("deleting last task clears focus", func(t *testing.T) {
		b := New()
		b.Add(NewDraft())
		b.Delete(1)
		if b.FocusID != 0 || b.SelID != 0 || len(b.Tasks) != 0 {
			t.Fatalf("focus=%d sel=%d len=%d", b.FocusID, b.SelID, len(b.Tasks))
		}
	})
}

func TestSetFocusAndNum(t *testing.T) {
	b := Demo()
	b.SelID = 4
	b.SetFocus()
	if b.FocusID != 4 {
		t.Fatalf("FocusID = %d", b.FocusID)
	}
	if b.Num(4) != 4 || b.Num(99) != 0 {
		t.Fatal("Num must be 1-based position")
	}
}

func TestTrackElapsed(t *testing.T) {
	b := Demo()
	b.TrackElapsed(1, "2026-10-22")
	if b.Get(1).Elapsed != 5048 {
		t.Fatalf("focused doing task elapsed = %d", b.Get(1).Elapsed)
	}
	b.FocusID = 3 // todo
	b.TrackElapsed(1, "2026-10-22")
	if b.Get(3).Elapsed != 0 || b.Get(3).Today != 0 {
		t.Fatal("todo focus must not accrue time")
	}
}

func TestTrackElapsedToday(t *testing.T) {
	b := Demo()
	b.TrackElapsed(1, "2026-10-22")
	b.TrackElapsed(1, "2026-10-22")
	got := b.Get(1)
	if got.Today != 2 || got.TodayDate != "2026-10-22" {
		t.Fatalf("today = %d on %q, want 2 on 2026-10-22", got.Today, got.TodayDate)
	}
	b.TrackElapsed(1, "2026-10-23")
	got = b.Get(1)
	if got.Today != 1 || got.TodayDate != "2026-10-23" {
		t.Fatalf("after rollover today = %d on %q, want 1 on 2026-10-23", got.Today, got.TodayDate)
	}
	if got.Elapsed != 5047+3 {
		t.Fatalf("rollover must keep elapsed, got %d", got.Elapsed)
	}
}

func TestStats(t *testing.T) {
	b := Demo()
	s := b.Stats()
	want := Stats{Doing: 2, Todo: 1, Done: 1, Elapsed: 5047 + 1820 + 2700}
	if s != want {
		t.Fatalf("Stats() = %+v, want %+v", s, want)
	}
}
