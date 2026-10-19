// Package notify alerts the user when a pomodoro phase ends: it rings the
// terminal bell and, when notify-send is available, shows a desktop
// notification.
package notify

import (
	"context"
	"io"
	"os"
	"os/exec"
	"time"
)

// sendTimeout bounds how long a notify-send call may take.
const sendTimeout = 5 * time.Second

// Notifier rings the bell on Bell and runs notify-send through LookPath/Run.
type Notifier struct {
	Bell     io.Writer
	LookPath func(file string) (string, error)
	Run      func(name string, args ...string) error
}

// New returns a Notifier that rings the bell on stderr and runs the real
// notify-send binary.
func New() Notifier {
	return Notifier{Bell: os.Stderr, LookPath: exec.LookPath, Run: run}
}

// Alert rings the bell and sends a desktop notification. Failures are
// ignored: an alert must never interrupt the app.
func (n Notifier) Alert(title, body string) {
	if n.Bell != nil {
		_, _ = io.WriteString(n.Bell, "\a")
	}
	if n.LookPath == nil || n.Run == nil {
		return
	}
	path, err := n.LookPath("notify-send")
	if err != nil {
		return
	}
	_ = n.Run(path, "--app-name=task_prioritizer", title, body)
}

func run(name string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), sendTimeout)
	defer cancel()
	return exec.CommandContext(ctx, name, args...).Run()
}
