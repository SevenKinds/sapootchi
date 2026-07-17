package game

import (
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/ui"
)

// EvolveModal is the evolution moment: the baby shakes with light, then grows
// into its next stage — full size from here on. Triggered from Home when the
// sim says the pet is ready.
type EvolveModal struct {
	blurUnder
	tick    int
	evolved bool
	from    string
}

// evolveShakeTicks: how long the "something is happening" beat lasts.
const evolveShakeTicks = 110

func NewEvolveModal(under Scene) *EvolveModal {
	return &EvolveModal{blurUnder: blurUnder{under: under}}
}

func (m *EvolveModal) Update(g *Game) error {
	m.tick++
	if m.tick == 1 {
		m.from = g.Pet.Phase.String()
	}
	// The reveal: apply the evolution exactly once, mid-modal.
	if !m.evolved && m.tick >= evolveShakeTicks {
		g.Pet.Evolve(time.Now())
		m.evolved = true
		g.Save()
	}
	if m.evolved && m.tick > evolveShakeTicks+30 && ui.Tapped(0, 0, ScreenW, ScreenH) {
		g.Pop()
	}
	return nil
}

func (m *EvolveModal) Draw(g *Game, screen *ebiten.Image) {
	m.blurUnder.draw(g, screen)

	const cw, ch = 300.0, 340.0
	x := (ScreenW - cw) / 2
	y := (ScreenH - ch) / 2
	drawModalCard(screen, x, y, cw, ch)

	img := g.petSprite()
	iw := float64(img.Bounds().Dx())
	ih := float64(img.Bounds().Dy())

	if !m.evolved {
		// Building up: small (baby scale) and shaking harder and harder.
		t := float64(m.tick) / evolveShakeTicks
		shake := math.Sin(float64(m.tick)*1.7) * 6 * t
		s := (150 / iw) * 0.60
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(s*ui.Scale, s*ui.Scale)
		op.GeoM.Translate((ScreenW/2-iw*s/2+shake)*ui.Scale, (y+150-ih*s/2)*ui.Scale)
		op.Filter = ebiten.FilterLinear
		screen.DrawImage(img, op)

		ui.DrawTextCenter(screen, g.Pet.Name+" is evolving!", ScreenW/2, y+26, 18, ui.Text, true)
		ui.DrawTextCenter(screen, "something is happening...", ScreenW/2, y+ch-44, 12, ui.TextDim, false)
		return
	}

	// The reveal: full size, star sparkles.
	ui.DrawImageFit(screen, img, x+40, y+58, cw-80, 180)
	for i := 0; i < 6; i++ {
		a := float64(m.tick)/14 + float64(i)*math.Pi/3
		sx := ScreenW/2 + math.Cos(a)*118
		sy := y + 150 + math.Sin(a)*86
		ui.DrawGlyph(screen, '\uf005', sx, sy, 11, ui.Gold) // star
	}
	ui.DrawTextCenter(screen, m.from+" -> "+g.Pet.Phase.String()+"!", ScreenW/2, y+26, 18, ui.Text, true)
	ui.DrawTextCenter(screen, g.Pet.Name+" grew up — full size!", ScreenW/2, y+ch-62, 12, ui.TextDim, false)
	ui.DrawTextCenter(screen, "tap to continue", ScreenW/2, y+ch-38, 11, ui.TextDim, false)
}
