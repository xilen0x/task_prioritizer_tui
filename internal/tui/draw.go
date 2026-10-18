package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/xilenox/tareas/internal/tui/theme"
)

// Low-level drawing helpers. Every visible cell is painted with an
// explicit background so the screen keeps the design's ink colour
// regardless of the terminal theme.

// st is a style with foreground fg on background bg.
func st(fg, bg string) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(fg)).Background(lipgloss.Color(bg))
}

// paint renders s with fg on bg.
func paint(s, fg, bg string) string { return st(fg, bg).Render(s) }

// bold renders s in bold with fg on bg.
func bold(s, fg, bg string) string { return st(fg, bg).Bold(true).Render(s) }

// blank is n background-coloured spaces.
func blank(n int, bg string) string {
	if n <= 0 {
		return ""
	}
	return lipgloss.NewStyle().Background(lipgloss.Color(bg)).Render(strings.Repeat(" ", n))
}

// fit pads s with background spaces, or truncates it, to exactly w cells.
func fit(s string, w int, bg string) string {
	sw := lipgloss.Width(s)
	if sw > w {
		return ansi.Truncate(s, w, "")
	}
	return s + blank(w-sw, bg)
}

// fitRight right-aligns s in w cells.
func fitRight(s string, w int, bg string) string {
	sw := lipgloss.Width(s)
	if sw > w {
		return ansi.Truncate(s, w, "")
	}
	return blank(w-sw, bg) + s
}

// center centres s in w cells.
func center(s string, w int, bg string) string {
	sw := lipgloss.Width(s)
	if sw >= w {
		return fit(s, w, bg)
	}
	l := (w - sw) / 2
	return blank(l, bg) + s + blank(w-sw-l, bg)
}

// ellipsis shortens plain text to w cells, ending in "…" when cut.
func ellipsis(s string, w int) string {
	if lipgloss.Width(s) <= w {
		return s
	}
	if w <= 1 {
		return ansi.Truncate(s, w, "")
	}
	return ansi.Truncate(s, w, "…")
}

// wrapLines word-wraps plain text into at most n lines of width w; the
// last line gets an ellipsis when text is left over.
func wrapLines(s string, w, n int) []string {
	words := strings.Fields(s)
	var lines []string
	cur := ""
	for i, word := range words {
		next := word
		if cur != "" {
			next = cur + " " + word
		}
		if lipgloss.Width(next) <= w {
			cur = next
			continue
		}
		if cur != "" {
			lines = append(lines, cur)
		}
		cur = word
		if len(lines) == n-1 {
			cur = strings.Join(words[i:], " ")
			break
		}
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	if len(lines) > n {
		lines = lines[:n]
	}
	for i := range lines {
		lines[i] = ellipsis(lines[i], w)
	}
	return lines
}

// spaceAround places items in h rows like CSS justify-content:
// space-around, returning h rows of width w.
func spaceAround(items []string, h, w int, bg string) []string {
	out := make([]string, h)
	for i := range out {
		out[i] = blank(w, bg)
	}
	n := len(items)
	if n == 0 {
		return out
	}
	free := float64(max(0, h-n))
	for i, it := range items {
		pos := int(free/float64(n)*(float64(i)+0.5)+0.5) + i
		if pos < h {
			out[pos] = fit(it, w, bg)
		}
	}
	return out
}

// boxSpec describes a framed panel with an inset title.
type boxSpec struct {
	width   int    // total width including borders
	title   string // inset title, may be empty
	titleFg string
	border  string // border colour
	dashed  bool
	bg      string // inner background
	padX    int    // inner horizontal padding
}

// box frames content lines. Each line is fitted to the inner width.
func box(b boxSpec, content []string) []string {
	h, v := "─", "│"
	tl, tr, bl, br := "╭", "╮", "╰", "╯"
	if b.dashed {
		h, v = "╌", "╎"
	}
	inner := b.width - 2
	top := paint(tl+h, b.border, theme.Ink1)
	used := 1
	if b.title != "" {
		t := " " + b.title + " "
		top += bold(t, b.titleFg, theme.Ink1)
		used += lipgloss.Width(t)
	}
	top += paint(strings.Repeat(h, max(0, inner-used))+tr, b.border, theme.Ink1)
	out := []string{top}
	side := paint(v, b.border, theme.Ink1)
	for _, c := range content {
		out = append(out, side+blank(b.padX, b.bg)+fit(c, inner-2*b.padX, b.bg)+blank(b.padX, b.bg)+side)
	}
	out = append(out, paint(bl+strings.Repeat(h, inner)+br, b.border, theme.Ink1))
	return out
}

// hjoin places blocks of lines side by side; all blocks must have the
// same number of lines (shorter ones are padded with background).
func hjoin(blocks ...[]string) []string {
	n := 0
	for _, b := range blocks {
		n = max(n, len(b))
	}
	out := make([]string, n)
	for _, b := range blocks {
		w := 0
		if len(b) > 0 {
			w = lipgloss.Width(b[0])
		}
		for i := 0; i < n; i++ {
			if i < len(b) {
				out[i] += b[i]
			} else {
				out[i] += blank(w, theme.Ink1)
			}
		}
	}
	return out
}

// column is a block of h background lines of width w.
func column(h, w int) []string {
	out := make([]string, h)
	for i := range out {
		out[i] = blank(w, theme.Ink1)
	}
	return out
}
