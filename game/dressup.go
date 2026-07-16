package game

import (
	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/ui"
)

// DressPage is the skins/cosmetics tab. POC stub: it previews the pet and
// placeholder cosmetic slots. Equipping/skins wire in here later.
type DressPage struct{}

func (p *DressPage) Icon() ui.Icon { return ui.IconDress }
func (p *DressPage) Label() string { return "Dress" }

func (p *DressPage) Update(g *Game) error { return nil }

func (p *DressPage) Draw(g *Game, screen *ebiten.Image) {
	ui.DrawTextBold(screen, "Dress Up", 24, 28, 24, ui.Text)
	g.DrawBlob(screen, ScreenW/2, 210)

	ui.FillRoundRect(screen, 16, 340, ScreenW-32, 150, 14, ui.Panel)
	ui.DrawTextBold(screen, "Cosmetics (coming soon)", 32, 356, 14, ui.Text)
	for i, slot := range []string{"Hat", "Accessory", "Skin"} {
		y := float32(386 + i*32)
		ui.FillRoundRect(screen, 32, y, 40, 24, 8, ui.PanelHi)
		ui.DrawText(screen, slot, 84, float64(y)+4, 13, ui.TextDim)
	}
}
