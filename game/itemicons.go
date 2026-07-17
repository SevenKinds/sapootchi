package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/simulation"
	"sapootchi/ui"
)

// All items draw as PIXEL SPRITES: pack art (brackeys fruit, Tiny Swords
// steak) plus code-drawn coffee/pill (pixelart.go). Glyphs remain only as a
// fallback.
var itemSprites map[simulation.FoodKind]*ebiten.Image

var itemGlyphs = map[simulation.FoodKind]struct {
	glyph rune
	clr   color.RGBA
}{
	simulation.FoodCoffee:     {'\U000F0176', color.RGBA{0xf4, 0xf0, 0xe8, 0xff}}, // md coffee
	simulation.FoodEnergyPill: {'\U000F0476', color.RGBA{0x46, 0xc7, 0xe0, 0xff}}, // md pill
}

// drawItemIcon renders an item centered at (cx, cy) with the given size.
func drawItemIcon(dst *ebiten.Image, kind simulation.FoodKind, cx, cy, size float64) {
	if spr, ok := itemSprites[kind]; ok {
		// Pixel art: nearest-neighbor, sized to the box. The meat sprite has
		// big transparent margins, so fit generously.
		f := size * 2.1 / float64(spr.Bounds().Dx())
		if spr.Bounds().Dx() <= 16 { // tiny fruit tiles fill the box directly
			f = size * 1.35 / 16
		}
		w := float64(spr.Bounds().Dx()) * f
		ui.DrawImageNearest(dst, spr, cx-w/2, cy-w/2, f, 1)
		return
	}
	if g, ok := itemGlyphs[kind]; ok {
		ui.DrawGlyph(dst, g.glyph, cx, cy, size, g.clr)
	}
}
