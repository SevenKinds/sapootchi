package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/simulation"
	"sapootchi/ui"
)

// Vector item icons (the Go font has no emoji). Each draws centered at (cx, cy)
// in roughly a 44x44 design-space box. When brand food art lands, these swap
// for sprites behind the same call.

func drawItemIcon(dst *ebiten.Image, kind simulation.FoodKind, cx, cy float64) {
	x, y := float32(cx), float32(cy)
	switch kind {
	case simulation.FoodApple:
		ui.FillCircle(dst, x, y+3, 16, color.RGBA{0xe6, 0x56, 0x4a, 0xff})
		ui.FillCircle(dst, x-6, y-2, 5, color.RGBA{0xff, 0x8c, 0x82, 0xff}) // shine
		ui.StrokeLine(dst, x, y-12, x+2, y-19, 3, color.RGBA{0x7a, 0x58, 0x38, 0xff})
		ui.FillCircle(dst, x+8, y-16, 5, ui.Good) // leaf
	case simulation.FoodSandwich:
		bread := color.RGBA{0xe8, 0xb8, 0x6d, 0xff}
		ui.FillRoundRect(dst, x-18, y-14, 36, 9, 4, bread)
		ui.FillRoundRect(dst, x-18, y-4, 36, 4, 2, ui.Good)                            // lettuce
		ui.FillRoundRect(dst, x-18, y+1, 36, 5, 2, color.RGBA{0xd9, 0x6c, 0x5c, 0xff}) // ham
		ui.FillRoundRect(dst, x-18, y+7, 36, 9, 4, bread)
	case simulation.FoodCake:
		ui.FillRoundRect(dst, x-17, y-2, 34, 18, 5, color.RGBA{0xf2, 0xd4, 0xa0, 0xff}) // sponge
		ui.FillRoundRect(dst, x-17, y-8, 34, 9, 4, color.RGBA{0xf7, 0xa8, 0xc4, 0xff})  // frosting
		ui.FillCircle(dst, x, y-13, 5, color.RGBA{0xd6, 0x3c, 0x4e, 0xff})              // cherry
	case simulation.FoodEnergyPill:
		// Capsule: cyan/white halves, slight tilt via offset halves.
		ui.FillRoundRect(dst, x-16, y-8, 32, 16, 8, ui.Energy)
		ui.FillRoundRect(dst, x-16, y-8, 16, 16, 8, color.RGBA{0xf4, 0xf0, 0xe8, 0xff})
		ui.FillCircle(dst, x-8, y-3, 2.5, color.RGBA{0xff, 0xff, 0xff, 0xff}) // shine
		// Little spark.
		ui.StrokeLine(dst, x+14, y-12, x+18, y-16, 2, ui.Gold)
		ui.StrokeLine(dst, x+18, y-12, x+14, y-16, 2, ui.Gold)
	case simulation.FoodCoffee:
		mug := color.RGBA{0xf4, 0xf0, 0xe8, 0xff}
		ui.FillRoundRect(dst, x-14, y-8, 26, 24, 5, mug)
		// Handle: short arc from segments.
		ui.StrokeLine(dst, x+12, y-3, x+18, y-1, 3, mug)
		ui.StrokeLine(dst, x+18, y-1, x+18, y+7, 3, mug)
		ui.StrokeLine(dst, x+18, y+7, x+12, y+9, 3, mug)
		ui.FillRoundRect(dst, x-11, y-5, 20, 7, 3, color.RGBA{0x6f, 0x4a, 0x2f, 0xff}) // coffee
		// Steam.
		ui.StrokeLine(dst, x-6, y-13, x-4, y-19, 2, ui.TextDim)
		ui.StrokeLine(dst, x+3, y-13, x+5, y-19, 2, ui.TextDim)
	}
}
