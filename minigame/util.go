package minigame

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/ui"
)

// imageRect is a tiny alias to keep sheet-slicing call sites short.
func imageRect(x0, y0, x1, y1 int) image.Rectangle {
	return image.Rect(x0, y0, x1, y1)
}

func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// drawSpinCoin renders the game's animated coin centered at (x, y) with the
// given design size — pickups ARE game coins. Falls back to a gold dot.
func drawSpinCoin(dst *ebiten.Image, frames []*ebiten.Image, ticks int, x, y, size float64) {
	if len(frames) == 0 {
		ui.FillCircle(dst, float32(x), float32(y), float32(size/2), ui.Gold)
		return
	}
	f := frames[(ticks/6)%len(frames)]
	sc := size / float64(f.Bounds().Dx())
	ui.DrawImageNearest(dst, f, x-size/2, y-size/2, sc, 1)
}

var _ = color.RGBA{} // keep color import if unused elsewhere
