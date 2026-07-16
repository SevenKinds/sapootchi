package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Icons for the tab bar, drawn from primitives (the Go font has no emoji).
// Each draws centered at (cx, cy) within roughly a 24x24 design-space box.

type Icon int

const (
	IconHome Icon = iota
	IconGames
	IconBag
	IconShop
	IconDress
	IconGear
)

// DrawIcon renders the icon centered at (cx, cy) in the given color.
func DrawIcon(dst *ebiten.Image, ic Icon, cx, cy float64, clr color.Color) {
	x, y := float32(cx), float32(cy)
	switch ic {
	case IconHome:
		// Roof + body.
		StrokeLine(dst, x-11, y, x, y-10, 2.5, clr)
		StrokeLine(dst, x, y-10, x+11, y, 2.5, clr)
		FillRoundRect(dst, x-8, y-2, 16, 12, 2, clr)
	case IconGames:
		// Gamepad: rounded body, d-pad cross, one button.
		FillRoundRect(dst, x-12, y-7, 24, 14, 7, clr)
		// carve is overdraw-free: draw details in bg color on top
		StrokeLine(dst, x-7, y, x-2, y, 2.2, Track)
		StrokeLine(dst, x-4.5, y-2.5, x-4.5, y+2.5, 2.2, Track)
		FillCircle(dst, x+6, y, 2.2, Track)
	case IconBag:
		// Satchel: body + handle.
		FillRoundRect(dst, x-9, y-4, 18, 13, 4, clr)
		StrokeLine(dst, x-5, y-4, x-5, y-9, 2.2, clr)
		StrokeLine(dst, x+5, y-4, x+5, y-9, 2.2, clr)
		StrokeLine(dst, x-5, y-9, x+5, y-9, 2.2, clr)
	case IconShop:
		// Price tag.
		FillRoundRect(dst, x-10, y-8, 15, 15, 3, clr)
		StrokeLine(dst, x+5, y-8, x+11, y-1, 3, clr)
		FillCircle(dst, x-5, y-3, 2, Track)
	case IconDress:
		// T-shirt: body + sleeves.
		FillRoundRect(dst, x-6, y-6, 12, 15, 2, clr)
		StrokeLine(dst, x-6, y-5, x-12, y-1, 4, clr)
		StrokeLine(dst, x+6, y-5, x+12, y-1, 4, clr)
	case IconGear:
		// Gear: ring + teeth.
		FillCircle(dst, x, y, 8, clr)
		FillCircle(dst, x, y, 3.5, Track)
		for _, d := range [][2]float32{{0, -11}, {0, 11}, {-11, 0}, {11, 0}, {-8, -8}, {8, -8}, {-8, 8}, {8, 8}} {
			FillCircle(dst, x+d[0]*0.85, y+d[1]*0.85, 2.4, clr)
		}
	}
}
