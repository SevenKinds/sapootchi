package game

import (
	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/ui"
)

// AnimModal plays a brand reaction animation over the game: the scene beneath
// keeps rendering softly blurred, and the card blends into the page gradient —
// the animation floats in the app world instead of feeling like a video
// player. (The animations show the classic SAPO regardless of the equipped
// skin — which is exactly why they live in a modal.)
type AnimModal struct {
	blurUnder
	frames []*ebiten.Image
	tick   int
}

// animModalFPS matches the extraction rate in cmd/assetprep.
const animModalFPS = 10

func NewAnimModal(frames []*ebiten.Image, under Scene) *AnimModal {
	return &AnimModal{blurUnder: blurUnder{under: under}, frames: frames}
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
	m.blurUnder.draw(g, screen)

	const cw, ch = 300.0, 320.0
	x := (ScreenW - cw) / 2
	y := (ScreenH - ch) / 2
	drawModalCard(screen, x, y, cw, ch)

	// Current frame, looping at extraction rate (ticks are 60/s).
	idx := (m.tick * animModalFPS / 60) % len(m.frames)
	ui.DrawImageFit(screen, m.frames[idx], x+20, y+24, cw-40, ch-72)

	ui.DrawTextCenter(screen, "tap to close", ScreenW/2, y+ch-32, 11, ui.TextDim, false)
}
