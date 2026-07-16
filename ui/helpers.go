package ui

import (
	"strconv"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// Itoa is exported for scenes that build small numeric strings.
func Itoa(n int) string { return strconv.Itoa(n) }

func itoa(n int) string { return strconv.Itoa(n) }

// measure returns the pixel width of s in the given face.
func measure(s string, f *text.GoTextFace) float64 {
	w, _ := text.Measure(s, f, f.Size*1.2)
	return w
}

// TextWidth returns the pixel width of s at the given size/weight.
func TextWidth(s string, size float64, isBold bool) float64 {
	return measure(s, face(size, isBold))
}
