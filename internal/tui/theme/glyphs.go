package theme

import (
	"math"
	"strings"
)

// glyphs is the 3×5 big-digit font; each cell is drawn two columns wide.
var glyphs = map[rune][5]string{
	'0': {"███", "█ █", "█ █", "█ █", "███"},
	'1': {"██ ", " █ ", " █ ", " █ ", "███"},
	'2': {"███", "  █", "███", "█  ", "███"},
	'3': {"███", "  █", "███", "  █", "███"},
	'4': {"█ █", "█ █", "███", "  █", "  █"},
	'5': {"███", "█  ", "███", "  █", "███"},
	'6': {"███", "█  ", "███", "█ █", "███"},
	'7': {"███", "  █", "  █", "  █", "  █"},
	'8': {"███", "█ █", "███", "█ █", "███"},
	'9': {"███", "█ █", "███", "  █", "███"},
	':': {" ", "█", " ", "█", " "},
	'%': {"█ █", "  █", " █ ", "█  ", "█ █"},
}

// BigHeight is the number of rows Big returns.
const BigHeight = 5

// Big renders s in the big-digit font as five rows. Glyphs are separated
// by one space; unknown runes render as 0, like the design.
func Big(s string) []string {
	rows := make([]strings.Builder, BigHeight)
	for i, c := range []rune(s) {
		g, ok := glyphs[c]
		if !ok {
			g = glyphs['0']
		}
		for r := 0; r < BigHeight; r++ {
			if i > 0 {
				rows[r].WriteByte(' ')
			}
			for _, cell := range g[r] {
				if cell == '█' {
					rows[r].WriteString("██")
				} else {
					rows[r].WriteString("  ")
				}
			}
		}
	}
	out := make([]string, BigHeight)
	for i := range rows {
		out[i] = rows[i].String()
	}
	return out
}

// BigWidth is the column width of Big(s).
func BigWidth(s string) int {
	w := 0
	for i, c := range []rune(s) {
		g, ok := glyphs[c]
		if !ok {
			g = glyphs['0']
		}
		if i > 0 {
			w++
		}
		w += 2 * len([]rune(g[0]))
	}
	return w
}

// BarStyle selects the glyph pair for progress bars.
type BarStyle string

// Bar styles (the design's barStyle prop).
const (
	Bloques BarStyle = "bloques"
	ASCII   BarStyle = "ascii"
	Fino    BarStyle = "fino"
)

var barGlyphs = map[BarStyle][2]string{
	Bloques: {"█", "░"},
	ASCII:   {"#", "·"},
	Fino:    {"━", "─"},
}

// ParseBarStyle parses a flag value, falling back to Bloques.
func ParseBarStyle(s string) (BarStyle, bool) {
	if _, ok := barGlyphs[BarStyle(s)]; ok {
		return BarStyle(s), true
	}
	return Bloques, false
}

func glyphPair(st BarStyle) (on, off string) {
	g, ok := barGlyphs[st]
	if !ok {
		g = barGlyphs[Bloques]
	}
	return g[0], g[1]
}

func filled(p, w int) int {
	n := int(math.Round(float64(p) / 100 * float64(w)))
	return max(0, min(w, n))
}

// Bar returns the filled and empty parts of a horizontal bar w cells wide.
func Bar(st BarStyle, p, w int) (fill, empty string) {
	on, off := glyphPair(st)
	n := filled(p, w)
	return strings.Repeat(on, n), strings.Repeat(off, w-n)
}

// Column returns the rows of a vertical bar h cells tall and three
// columns wide: empty rows (top) then filled rows (bottom).
func Column(st BarStyle, p, h int) (fill, empty []string) {
	on, off := glyphPair(st)
	n := filled(p, h)
	for i := 0; i < h-n; i++ {
		empty = append(empty, strings.Repeat(off, 3))
	}
	for i := 0; i < n; i++ {
		fill = append(fill, strings.Repeat(on, 3))
	}
	return fill, empty
}
