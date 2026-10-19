package theme

import (
	"fmt"
	"strconv"
)

// HMS formats seconds as hh:mm:ss.
func HMS(s int) string { return fmt.Sprintf("%02d:%02d:%02d", s/3600, s/60%60, s%60) }

// HoursLabel formats an estimate: "6 h", "1.5 h".
func HoursLabel(h float64) string {
	if h == float64(int(h)) {
		return strconv.Itoa(int(h)) + " h"
	}
	return strconv.FormatFloat(h, 'f', 1, 64) + " h"
}

// TotalLabel formats total tracked seconds as h:mm.
func TotalLabel(s int) string { return fmt.Sprintf("%d:%02d", s/3600, s/60%60) }
