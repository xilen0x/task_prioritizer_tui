package theme

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

func golden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	if *update {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden (run with -update): %v", err)
	}
	if got != string(want) {
		t.Fatalf("%s mismatch\n--- got ---\n%s\n--- want ---\n%s", name, got, want)
	}
}

func TestBigDigitsGolden(t *testing.T) {
	var sb strings.Builder
	for _, s := range []string{"12:34", "70%", "0123456789", "100%"} {
		sb.WriteString(s + "\n" + strings.Join(Big(s), "\n") + "\n\n")
	}
	golden(t, "big", sb.String())
}

func TestBigWidth(t *testing.T) {
	tests := map[string]int{"70%": 20, "17:43": 30, "100%": 27}
	for in, want := range tests {
		rows := Big(in)
		if len(rows) != 5 {
			t.Fatalf("Big(%q) has %d rows", in, len(rows))
		}
		for _, r := range rows {
			if got := len([]rune(r)); got != want {
				t.Fatalf("Big(%q) row width %d, want %d", in, got, want)
			}
		}
		if BigWidth(in) != want {
			t.Fatalf("BigWidth(%q) = %d, want %d", in, BigWidth(in), want)
		}
	}
}

func TestBarsGolden(t *testing.T) {
	var sb strings.Builder
	for _, st := range []BarStyle{Bloques, ASCII, Fino} {
		for _, p := range []int{0, 40, 70, 100} {
			f, e := Bar(st, p, 18)
			sb.WriteString(string(st) + " " + f + e + "\n")
		}
		fill, empty := Column(st, 40, 13)
		sb.WriteString(strings.Join(empty, "\n") + "\n" + strings.Join(fill, "\n") + "\n\n")
	}
	golden(t, "bars", sb.String())
}

func TestBarRounding(t *testing.T) {
	f, e := Bar(Bloques, 70, 62)
	if n := len([]rune(f)); n != 43 {
		t.Fatalf("fill = %d cells, want 43", n)
	}
	if n := len([]rune(f + e)); n != 62 {
		t.Fatalf("bar width = %d, want 62", n)
	}
}

func TestParseBarStyle(t *testing.T) {
	if s, ok := ParseBarStyle("fino"); !ok || s != Fino {
		t.Fatal("fino not parsed")
	}
	if s, ok := ParseBarStyle("nope"); ok || s != Bloques {
		t.Fatal("unknown style must fall back to bloques")
	}
}

func TestMix(t *testing.T) {
	tests := []struct {
		alpha float64
		want  string
	}{{1, "#FFFFFF"}, {0.85, "#DEDFE0"}, {0.65, "#B2B4B7"}, {0.45, "#85898E"}}
	for _, tt := range tests {
		if got := mix("#FFFFFF", "#222831", tt.alpha); got != tt.want {
			t.Fatalf("mix(%v) = %s, want %s", tt.alpha, got, tt.want)
		}
	}
}

func TestPctColor(t *testing.T) {
	tests := []struct {
		p    int
		want string
	}{{0, OnDark3}, {10, "#F0524A"}, {25, "#F0524A"}, {40, "#F5922A"}, {70, "#FFFE5A"}, {90, "#A6E39A"}, {100, "#2BD46A"}}
	for _, tt := range tests {
		if got := string(PctColor(tt.p)); got != tt.want {
			t.Fatalf("PctColor(%d) = %s, want %s", tt.p, got, tt.want)
		}
	}
}

func TestFormat(t *testing.T) {
	if got := HMS(5047); got != "01:24:07" {
		t.Fatalf("HMS = %q", got)
	}
	if got := HoursLabel(6); got != "6 h" {
		t.Fatalf("HoursLabel(6) = %q", got)
	}
	if got := HoursLabel(1.5); got != "1,5 h" {
		t.Fatalf("HoursLabel(1.5) = %q", got)
	}
	if got := TotalLabel(5047 + 1820 + 2700); got != "2:39" {
		t.Fatalf("TotalLabel = %q", got)
	}
}
