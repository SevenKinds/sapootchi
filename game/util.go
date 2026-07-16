package game

import "image/color"

// colIf picks a or b by condition — handy for enabled/disabled text colors.
func colIf(cond bool, a, b color.Color) color.Color {
	if cond {
		return a
	}
	return b
}
