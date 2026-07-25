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

// drawBlobStandin draws the placeholder pet used when the real-sprite setting is
// off: a rounded green body with two googly eyes, so the stand-in still reads as
// the blob (some games used to draw an eyeless blob). (x,y) is the box top-left;
// w,h its size — eyes scale with the box, so squash/stretch is preserved.
func drawBlobStandin(dst *ebiten.Image, x, y, w, h float64) {
	rad := w * 0.42
	if h*0.42 < rad {
		rad = h * 0.42
	}
	ui.FillRoundRect(dst, float32(x), float32(y), float32(w), float32(h), float32(rad), ui.Good)

	er := h * 0.15 // eye radius
	pr := er * 0.42
	ey := y + h*0.42
	for _, ex := range []float64{x + w*0.36, x + w*0.64} {
		ui.FillCircle(dst, float32(ex), float32(ey), float32(er), color.White)
		ui.FillCircle(dst, float32(ex+er*0.12), float32(ey+er*0.10), float32(pr), color.RGBA{0x12, 0x16, 0x1c, 0xff})
	}
}
