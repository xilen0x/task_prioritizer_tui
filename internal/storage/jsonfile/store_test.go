package jsonfile

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/xilenox/tareas/internal/pomodoro"
	"github.com/xilenox/tareas/internal/tasks"
)

func TestRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "state.json")
	s := New(path)
	want := State{Board: tasks.Demo(), Pomo: pomodoro.Timer{Phase: pomodoro.Descanso, Left: 120, Done: 2}}
	if err := s.Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, found, err := s.Load()
	if err != nil || !found {
		t.Fatalf("Load: found=%v err=%v", found, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip mismatch\n got %+v\nwant %+v", got, want)
	}
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 {
		t.Fatalf("temp files left behind: %v", entries)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, found, err := New(filepath.Join(t.TempDir(), "none.json")).Load()
	if err != nil || found {
		t.Fatalf("found=%v err=%v, want not found and no error", found, err)
	}
}

func TestLoadCorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := New(path).Load(); err == nil {
		t.Fatal("want error for corrupt file")
	}
}

func TestLoadRepairsInconsistentState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	raw := `{"version":1,"board":{"tasks":[
		{"id":7,"title":"a","status":"curso","prio":9,"progress":135,"est":0,"elapsed":-3},
		{"id":8,"title":"b","status":"rara","prio":2,"progress":10,"est":1,"elapsed":0},
		{"id":9,"title":"c","status":"hecha","prio":2,"progress":100,"est":1,"elapsed":0},
		{"id":10,"title":"d","status":"hecha","prio":2,"progress":100,"est":1,"elapsed":0},
		{"id":11,"title":"e","status":"hecha","prio":2,"progress":100,"est":1,"elapsed":0}],
		"focusId":42,"selId":42,"nextId":1},
		"pomo":{"phase":"x","left":0,"running":true,"done":-1}}`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	st, _, err := New(path).Load()
	if err != nil {
		t.Fatal(err)
	}
	b := st.Board
	if len(b.Tasks) != tasks.MaxTasks {
		t.Fatalf("len = %d, want %d", len(b.Tasks), tasks.MaxTasks)
	}
	first := b.Tasks[0]
	if first.Prio != 3 || first.Progress != 100 || first.Est != tasks.EstMin || first.Elapsed != 0 {
		t.Fatalf("first task not clamped: %+v", first)
	}
	if b.Tasks[1].Status != tasks.Pendiente {
		t.Fatalf("unknown status not repaired: %q", b.Tasks[1].Status)
	}
	if b.FocusID != 7 || b.SelID != 7 || b.NextID != 11 {
		t.Fatalf("focus=%d sel=%d next=%d", b.FocusID, b.SelID, b.NextID)
	}
	if st.Pomo != (pomodoro.Timer{Phase: pomodoro.Trabajo, Left: pomodoro.WorkSeconds, Running: true}) {
		t.Fatalf("pomodoro not repaired: %+v", st.Pomo)
	}
}

func TestDefaultPath(t *testing.T) {
	t.Run("uses XDG_DATA_HOME", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", "/x/data")
		got, err := DefaultPath()
		if err != nil || got != filepath.Join("/x/data", "tareas", "state.json") {
			t.Fatalf("got %q err=%v", got, err)
		}
	})
	t.Run("falls back to ~/.local/share", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("XDG_DATA_HOME", "")
		t.Setenv("HOME", home)
		got, err := DefaultPath()
		if err != nil || got != filepath.Join(home, ".local", "share", "tareas", "state.json") {
			t.Fatalf("got %q err=%v", got, err)
		}
	})
}
