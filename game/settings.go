package game

import (
	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/simulation"
	"sapootchi/ui"
)

// SettingsPage holds player preferences. Toggles persist with the save file.
type SettingsPage struct {
	flash      string
	flashUntil int
}

func (p *SettingsPage) Icon() ui.Icon { return ui.IconGear }
func (p *SettingsPage) Label() string { return "More" }

func (p *SettingsPage) Update(g *Game) error {
	if x, y, w, h := p.toggleRect(0); ui.Tapped(x, y, w, h) {
		g.Settings.RealSpriteInGames = !g.Settings.RealSpriteInGames
		g.Save()
	}
	if DevMode {
		for i, b := range p.devButtons() {
			if !b.Clicked() {
				continue
			}
			switch i {
			case 0:
				g.Pet.AddFood(simulation.FoodEnergyPill, 1)
				p.flash = "+1 Energy Pill (check Items)"
			case 1:
				g.Pet.Coins += 100
				p.flash = "+100 coins"
			}
			p.flashUntil = g.tick + 150
			g.Save()
		}
	}
	return nil
}

func (p *SettingsPage) Draw(g *Game, screen *ebiten.Image) {
	ui.DrawTextBold(screen, "Settings", 24, 28, 24, ui.Text)

	p.drawToggleRow(screen, 0,
		"Real pet in mini-games",
		"Play as your blob instead of a stand-in shape",
		g.Settings.RealSpriteInGames)

	if DevMode {
		y := float32(p.devY() - 26)
		ui.FillRoundRect(screen, 24, y, 44, 18, 6, ui.Bad)
		ui.DrawTextBold(screen, "DEV", 34, float64(y)+1, 11, ui.Text)
		for _, b := range p.devButtons() {
			b.Draw(screen, true)
		}
		if g.tick < p.flashUntil {
			ui.DrawTextCenter(screen, p.flash, ScreenW/2, p.devY()+56, 12, ui.Gold, true)
		}
	}
}

func (p *SettingsPage) devY() float64 { return 220 }

func (p *SettingsPage) devButtons() []ui.Button {
	y := p.devY()
	return []ui.Button{
		{X: 24, Y: y, W: (ScreenW - 56) / 2, H: 44, Label: "+1 Energy Pill"},
		{X: 32 + (ScreenW-56)/2, Y: y, W: (ScreenW - 56) / 2, H: 44, Label: "+100 coins"},
	}
}

func (p *SettingsPage) toggleRect(i int) (x, y, w, h float64) {
	return 24, float64(96 + i*88), ScreenW - 48, 72
}

func (p *SettingsPage) drawToggleRow(screen *ebiten.Image, i int, title, desc string, on bool) {
	x, y, w, h := p.toggleRect(i)
	ui.FillRoundRect(screen, float32(x), float32(y), float32(w), float32(h), 14, ui.Panel)
	ui.DrawTextBold(screen, title, x+18, y+14, 15, ui.Text)
	ui.DrawText(screen, desc, x+18, y+38, 11, ui.TextDim)

	// Switch: track + knob.
	swW, swH := 52.0, 28.0
	swX, swY := x+w-swW-18, y+(h-swH)/2
	track := ui.Track
	knobX := swX + 3
	if on {
		track = ui.Good
		knobX = swX + swW - swH + 3
	}
	ui.FillRoundRect(screen, float32(swX), float32(swY), float32(swW), float32(swH), float32(swH/2), track)
	ui.FillCircle(screen, float32(knobX+swH/2-3), float32(swY+swH/2), float32(swH/2-4), ui.Text)
}
