package tasks

import "testing"

func TestTodaySeconds(t *testing.T) {
	tests := []struct {
		name string
		task Task
		date string
		want int
	}{
		{"same day returns counter", Task{Today: 90, TodayDate: "2026-10-22"}, "2026-10-22", 90},
		{"stale day returns zero", Task{Today: 90, TodayDate: "2026-10-21"}, "2026-10-22", 0},
		{"never tracked returns zero", Task{}, "2026-10-22", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.task.TodaySeconds(tt.date); got != tt.want {
				t.Fatalf("TodaySeconds = %d, want %d", got, tt.want)
			}
		})
	}
}
