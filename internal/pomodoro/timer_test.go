package pomodoro

import "testing"

func TestNewStartsRunningWorkBlock(t *testing.T) {
	got := New()
	want := Timer{Phase: Work, Left: WorkSeconds, Running: true}
	if got != want {
		t.Fatalf("New() = %+v, want %+v", got, want)
	}
}

func TestTick(t *testing.T) {
	tests := []struct {
		name       string
		in         Timer
		want       Timer
		wantWorked bool
	}{
		{"paused does nothing", Timer{Phase: Work, Left: 10}, Timer{Phase: Work, Left: 10}, false},
		{"work counts down", Timer{Phase: Work, Left: 10, Running: true}, Timer{Phase: Work, Left: 9, Running: true}, true},
		{"break counts down without work", Timer{Phase: Break, Left: 10, Running: true}, Timer{Phase: Break, Left: 9, Running: true}, false},
		{"work end starts short break", Timer{Phase: Work, Left: 1, Running: true, Done: 0}, Timer{Phase: Break, Left: ShortBreakSeconds, Running: true, Done: 1}, true},
		{"fourth pomodoro starts long break", Timer{Phase: Work, Left: 1, Running: true, Done: 3}, Timer{Phase: Break, Left: LongBreakSeconds, Running: true, Done: 4}, true},
		{"eighth pomodoro starts long break", Timer{Phase: Work, Left: 1, Running: true, Done: 7}, Timer{Phase: Break, Left: LongBreakSeconds, Running: true, Done: 8}, true},
		{"break end starts work", Timer{Phase: Break, Left: 1, Running: true, Done: 4}, Timer{Phase: Work, Left: WorkSeconds, Running: true, Done: 4}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, worked := tt.in.Tick()
			if got != tt.want || worked != tt.wantWorked {
				t.Fatalf("Tick() = %+v,%v want %+v,%v", got, worked, tt.want, tt.wantWorked)
			}
		})
	}
}

func TestFullCycleOfTicks(t *testing.T) {
	tm := New()
	worked := 0
	for i := 0; i < WorkSeconds; i++ {
		var w bool
		tm, w = tm.Tick()
		if w {
			worked++
		}
	}
	if tm.Phase != Break || tm.Left != ShortBreakSeconds || tm.Done != 1 || worked != WorkSeconds {
		t.Fatalf("after one work block: %+v worked=%d", tm, worked)
	}
}

func TestToggleAndReset(t *testing.T) {
	tm := Timer{Phase: Break, Left: 42, Running: true, Done: 3}
	if tm.Toggled().Running {
		t.Fatal("Toggled must pause")
	}
	r := tm.Reset()
	want := Timer{Phase: Work, Left: WorkSeconds, Running: false, Done: 3}
	if r != want {
		t.Fatalf("Reset() = %+v, want %+v", r, want)
	}
}

func TestLabels(t *testing.T) {
	tm := Timer{Phase: Work, Left: 1500 - 437, Done: 6}
	if got := tm.MMSS(); got != "17:43" {
		t.Fatalf("MMSS() = %q", got)
	}
	if got := tm.CycleDone(); got != 2 {
		t.Fatalf("CycleDone() = %d", got)
	}
}
