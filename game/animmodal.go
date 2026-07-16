package game

import (
	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/ui"
)

// AnimModal plays a brand reaction animation over the game. The scene beneath
// keeps rendering, softly BLURRED, and the card is painted with the same
// background gradient as the page — so the animation floats in the app world
// instead of feeling like a video player. (The animations show the classic
// SAPO regardless of the equipped skin — which is exactly why they live in a
// modal instead of replacing the pet on screen.)
type AnimModal struct {
	frames []*ebiten.Image
	under  Scene // the scene rendered blurred behind the card
	tick   int

	underBuf *ebiten.Image // full-res capture of the scene below
	smallBuf *ebiten.Image // downscale buffer (the cheap blur)
}

// animModalFPS matches the extraction rate in cmd/assetprep.
const animModalFPS = 10

// blurFactor: downscale-then-upscale with linear filtering = a cheap soft blur.
const blurFactor = 7

func NewAnimModal(frames []*ebiten.Image, under Scene) *AnimModal {
	return &AnimModal{frames: frames, under: under}
}

func (m *AnimModal) Update(g *Game) error {
	m.tick++
	// Small grace so the opening tap doesn't immediately close it.
	if m.tick > 15 && ui.Tapped(0, 0, ScreenW, ScreenH) {
		g.Pop()
	}
	return nil
}

func (m *AnimModal) Draw(g *Game, screen *ebiten.Image) {
	m.drawBlurredUnder(g, screen)

	// Card: painted with the page's own gradient at its rows, so it blends
	// into the world; the soft shadow is what separates it from the blur.
	const cw, ch = 300.0, 320.0
	x := (ScreenW - cw) / 2
	y := (ScreenH - ch) / 2
	ui.FillRoundRect(screen, float32(x), float32(y+5), cw, ch, 20, ui.Shadow)
	// Rounded caps in their local gradient color, square bands between.
	ui.FillRoundRect(screen, float32(x), float32(y), cw, 40, 20, ui.BGColorAt(y+10, ScreenH))
	ui.FillRoundRect(screen, float32(x), float32(y+ch-40), cw, 40, 20, ui.BGColorAt(y+ch-10, ScreenH))
	const bands = 14
	inner := ch - 40
	for i := 0; i < bands; i++ {
		by := y + 20 + inner*float64(i)/bands
		ui.FillRoundRect(screen, float32(x), float32(by), cw, float32(inner/bands+1), 0,
			ui.BGColorAt(by, ScreenH))
	}

	// Current frame, looping at extraction rate (ticks are 60/s).
	idx := (m.tick * animModalFPS / 60) % len(m.frames)
	ui.DrawImageFit(screen, m.frames[idx], x+20, y+24, cw-40, ch-72)

	ui.DrawTextCenter(screen, "tap to close", ScreenW/2, y+ch-32, 11, ui.TextDim, false)
}

// drawBlurredUnder renders the scene below into a buffer, then downscales and
// upscales it with linear filtering — a cheap, soft, full-screen blur.
func (m *AnimModal) drawBlurredUnder(g *Game, screen *ebiten.Image) {
	fw, fh := screen.Bounds().Dx(), screen.Bounds().Dy()
	if m.underBuf == nil || m.underBuf.Bounds().Dx() != fw || m.underBuf.Bounds().Dy() != fh {
		m.underBuf = ebiten.NewImage(fw, fh)
		m.smallBuf = ebiten.NewImage(fw/blurFactor+1, fh/blurFactor+1)
	}

	m.underBuf.Clear()
	ui.BackgroundGradient(m.underBuf)
	m.under.Draw(g, m.underBuf)

	m.smallBuf.Clear()
	down := &ebiten.DrawImageOptions{}
	down.GeoM.Scale(1.0/blurFactor, 1.0/blurFactor)
	down.Filter = ebiten.FilterLinear
	m.smallBuf.DrawImage(m.underBuf, down)

	up := &ebiten.DrawImageOptions{}
	up.GeoM.Scale(blurFactor, blurFactor)
	up.Filter = ebiten.FilterLinear
	screen.DrawImage(m.smallBuf, up)
}
