// Package theme holds the tareas palette, bar glyph styles, the big-digit
// font and small presentation formatters.
package theme

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/lipgloss"
)

// Palette (hex). The on-dark text levels are white at 100/85/65/45% and
// the border/rule tones white at 20/12%, pre-mixed onto Ink1.
const (
	Ink1    = "#222831"
	Ink2    = "#30475E"
	Accent  = "#F05454"
	Success = "#22934F"
	Danger  = "#B4342B"
	// DangerBorder is the lighter danger tone used for the confirm title.
	DangerBorder = "#E0675E"
	White        = "#FFFFFF"

	OnDark1      = "#FFFFFF"
	OnDark2      = "#DEDFE0"
	OnDark3      = "#B2B4B7"
	OnDark4      = "#85898E"
	OnDarkBorder = "#4E535A"
	OnDarkRule   = "#3D424A"

	Doing = "#FFFE5A"
)

// C converts a hex constant to a lipgloss colour.
func C(hex string) lipgloss.Color { return lipgloss.Color(hex) }

// PctColor is the progress colour ramp of the design.
func PctColor(p int) lipgloss.Color {
	switch {
	case p <= 0:
		return OnDark3
	case p <= 25:
		return "#F0524A"
	case p <= 50:
		return "#F5922A"
	case p <= 75:
		return "#FFFE5A"
	case p < 100:
		return "#A6E39A"
	default:
		return "#2BD46A"
	}
}

// mix blends fg over bg with the given alpha, returning #RRGGBB. It
// documents how the on-dark levels above were derived.
func mix(fg, bg string, alpha float64) string {
	f, b := rgb(fg), rgb(bg)
	var out [3]int
	for i := range out {
		v := float64(f[i])*alpha + float64(b[i])*(1-alpha)
		out[i] = int(v + 0.5)
	}
	return fmt.Sprintf("#%02X%02X%02X", out[0], out[1], out[2])
}

func rgb(hex string) [3]int {
	n, _ := strconv.ParseUint(hex[1:], 16, 32)
	return [3]int{int(n >> 16 & 0xFF), int(n >> 8 & 0xFF), int(n & 0xFF)}
}
