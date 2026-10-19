package main

import (
	"path/filepath"
	"testing"

	"github.com/xilen0x/task_prioritizer_tui/internal/pomodoro"
	"github.com/xilen0x/task_prioritizer_tui/internal/storage/jsonfile"
	"github.com/xilen0x/task_prioritizer_tui/internal/tasks"
)

func TestParseFlags(t *testing.T) {
	o, err := parseFlags([]string{"--bar=thin", "--plain-digits", "--demo", "--no-alert"})
	if err != nil || o.bar != "thin" || !o.plainDigits || !o.demo || !o.noAlert {
		t.Fatalf("got %+v err=%v", o, err)
	}
	if _, err := parseFlags([]string{"--bar=x"}); err == nil {
		t.Fatal("want error for invalid --bar")
	}
	if o, err := parseFlags([]string{"--version"}); err != nil || !o.version {
		t.Fatalf("--version: got %+v err=%v", o, err)
	}
}

func TestInitialState(t *testing.T) {
	t.Run("demo seeds sample board when no state exists", func(t *testing.T) {
		st, err := initialState(jsonfile.New(filepath.Join(t.TempDir(), "s.json")), true)
		if err != nil || len(st.Board.Tasks) != 4 || st.Pomo.Left != 1063 || st.Pomo.Done != 2 {
			t.Fatalf("got %+v err=%v", st, err)
		}
	})
	t.Run("without demo starts empty", func(t *testing.T) {
		st, err := initialState(jsonfile.New(filepath.Join(t.TempDir(), "s.json")), false)
		if err != nil || len(st.Board.Tasks) != 0 || st.Pomo != pomodoro.New() {
			t.Fatalf("got %+v err=%v", st, err)
		}
	})
	t.Run("saved state wins over demo", func(t *testing.T) {
		store := jsonfile.New(filepath.Join(t.TempDir(), "s.json"))
		b := tasks.New()
		b.Add(tasks.Task{Title: "propia", Prio: 2, Est: 1})
		if err := store.Save(jsonfile.State{Board: b, Pomo: pomodoro.New()}); err != nil {
			t.Fatal(err)
		}
		st, err := initialState(store, true)
		if err != nil || len(st.Board.Tasks) != 1 || st.Board.Tasks[0].Title != "propia" {
			t.Fatalf("got %+v err=%v", st, err)
		}
	})
}
