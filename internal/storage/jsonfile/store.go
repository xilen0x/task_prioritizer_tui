// Package jsonfile persists the board and pomodoro state as a JSON file,
// written atomically (temp file + rename).
package jsonfile

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/xilen0x/task_prioritizer_tui/internal/pomodoro"
	"github.com/xilen0x/task_prioritizer_tui/internal/tasks"
)

const formatVersion = 1

// State is everything the app persists.
type State struct {
	Board tasks.Board    `json:"board"`
	Pomo  pomodoro.Timer `json:"pomo"`
}

type fileFormat struct {
	Version int `json:"version"`
	State
}

// Store reads and writes one state file.
type Store struct{ path string }

// New returns a store backed by path.
func New(path string) *Store { return &Store{path: path} }

// Path is the file the store uses.
func (s *Store) Path() string { return s.path }

// AppDir is the data directory name; legacyAppDir is the name used before
// the app was renamed.
const (
	AppDir       = "task_prioritizer"
	legacyAppDir = "tareas"
)

// DefaultPath is $XDG_DATA_HOME/task_prioritizer/state.json, falling back to
// ~/.local/share/task_prioritizer/state.json.
func DefaultPath() (string, error) {
	dir := os.Getenv("XDG_DATA_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("locate home directory: %w", err)
		}
		dir = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dir, AppDir, "state.json"), nil
}

// MigrateLegacy moves the state file from the old "tareas" data directory
// next to path's directory when path does not exist yet.
func MigrateLegacy(path string) error {
	if _, err := os.Stat(path); err == nil || !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	legacy := filepath.Join(filepath.Dir(filepath.Dir(path)), legacyAppDir, filepath.Base(path))
	if _, err := os.Stat(legacy); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}
	if err := os.Rename(legacy, path); err != nil {
		return fmt.Errorf("move legacy state: %w", err)
	}
	return nil
}

// Load reads the state. found is false (with a nil error) when the file
// does not exist yet. Loaded data is repaired to satisfy domain limits.
func (s *Store) Load() (st State, found bool, err error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return State{}, false, nil
	}
	if err != nil {
		return State{}, false, fmt.Errorf("read state: %w", err)
	}
	var f fileFormat
	if err := json.Unmarshal(data, &f); err != nil {
		return State{}, true, fmt.Errorf("parse state %s: %w", s.path, err)
	}
	return repair(f.State), true, nil
}

// Save writes the state atomically, creating the directory if needed.
func (s *Store) Save(st State) error {
	data, err := json.MarshalIndent(fileFormat{Version: formatVersion, State: st}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".state-*.json")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmp.Name()) // no-op after a successful rename
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmp.Name(), s.path); err != nil {
		return fmt.Errorf("replace state file: %w", err)
	}
	return nil
}

func clamp(v, lo, hi int) int { return max(lo, min(hi, v)) }

// Values written by the Spanish versions of the app, mapped on load so
// existing state files keep working.
var (
	legacyStatus = map[tasks.Status]tasks.Status{"pendiente": tasks.Todo, "curso": tasks.Doing, "hecha": tasks.Done}
	legacyPhase  = map[pomodoro.Phase]pomodoro.Phase{"trabajo": pomodoro.Work, "descanso": pomodoro.Break}
)

// repair makes a hand-edited or older file safe for the domain.
func repair(st State) State {
	b := st.Board
	if len(b.Tasks) > tasks.MaxTasks {
		b.Tasks = b.Tasks[:tasks.MaxTasks]
	}
	maxID := 0
	for i := range b.Tasks {
		t := &b.Tasks[i]
		if s, ok := legacyStatus[t.Status]; ok {
			t.Status = s
		}
		switch t.Status {
		case tasks.Todo, tasks.Doing, tasks.Done:
		default:
			t.Status = tasks.Todo
		}
		t.Prio = clamp(t.Prio, tasks.PrioMin, tasks.PrioMax)
		t.Progress = clamp(t.Progress, 0, 100)
		t.Est = max(tasks.EstMin, min(tasks.EstMax, t.Est))
		t.Elapsed = max(0, t.Elapsed)
		t.Today = max(0, t.Today)
		maxID = max(maxID, t.ID)
	}
	b.NextID = max(b.NextID, maxID+1)
	if !b.Has(b.FocusID) {
		b.FocusID = 0
		if len(b.Tasks) > 0 {
			b.FocusID = b.Tasks[0].ID
		}
	}
	if !b.Has(b.SelID) {
		b.SelID = b.FocusID
	}
	st.Board = b

	p := st.Pomo
	if ph, ok := legacyPhase[p.Phase]; ok {
		p.Phase = ph
	}
	if p.Done < 0 {
		p.Done = 0
	}
	switch {
	case p.Phase == pomodoro.Work && p.Left > 0 && p.Left <= pomodoro.WorkSeconds:
	case p.Phase == pomodoro.Break && p.Left > 0 && p.Left <= pomodoro.LongBreakSeconds:
	default:
		p.Phase, p.Left = pomodoro.Work, pomodoro.WorkSeconds
	}
	st.Pomo = p
	return st
}
