package notify

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
)

func TestAlertRingsBellAndSendsNotification(t *testing.T) {
	var out bytes.Buffer
	var ran []string
	n := Notifier{
		Bell:     &out,
		LookPath: func(string) (string, error) { return "/usr/bin/notify-send", nil },
		Run:      func(name string, args ...string) error { ran = append([]string{name}, args...); return nil },
	}
	n.Alert("Pomodoro done", "Take a 5 min break.")
	if out.String() != "\a" {
		t.Fatalf("bell = %q", out.String())
	}
	want := []string{"/usr/bin/notify-send", "--app-name=task_prioritizer", "Pomodoro done", "Take a 5 min break."}
	if !reflect.DeepEqual(ran, want) {
		t.Fatalf("ran %v, want %v", ran, want)
	}
}

func TestAlertSkipsMissingNotifySend(t *testing.T) {
	var out bytes.Buffer
	n := Notifier{
		Bell:     &out,
		LookPath: func(string) (string, error) { return "", errors.New("not found") },
		Run:      func(string, ...string) error { t.Fatal("must not run"); return nil },
	}
	n.Alert("a", "b")
	if out.String() != "\a" {
		t.Fatalf("bell = %q", out.String())
	}
}
