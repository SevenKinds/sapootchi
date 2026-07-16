package game

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/ui"
)

// DressPage is the skins tab: a grid of the brand's themed SAPO looks. Tap a
// tile to equip; "Classic" is the default blob. The equipped skin persists.
type DressPage struct {
	flash      string
	flashUntil int
}

func (p *DressPage) Icon() ui.Icon { return ui.IconDress }
func (p *DressPage) Label() string { return "Dress" }

// Grid layout (design-space).
const (
	dressGridY = 236.0
	dressCols  = 4
	dressTile  = 74.0
	dressGap   = 8.0
)

// skinAt maps tile index -> skin name ("" = classic, index 0).
func (p *DressPage) skinAt(g *Game, i int) (name string, ok bool) {
	if i == 0 {
		return "", true
	}
	if i-1 < len(g.Sprites.SkinNames) {
		return g.Sprites.SkinNames[i-1], true
	}
	return "", false
}

func (p *DressPage) tileRect(i int) (x, y, w, h float64) {
	total := dressCols*dressTile + (dressCols-1)*dressGap
	x0 := (ScreenW - total) / 2
	col, row := i%dressCols, i/dressCols
	return x0 + float64(col)*(dressTile+dressGap), dressGridY + float64(row)*(dressTile+dressGap), dressTile, dressTile
}

func (p *DressPage) Update(g *Game) error {
	n := 1 + len(g.Sprites.SkinNames)
	for i := 0; i < n; i++ {
		if !ui.Tapped(p.tileRect(i)) {
			continue
		}
		name, ok := p.skinAt(g, i)
		if !ok || g.Pet.Skin == name {
			continue
		}
		g.Pet.Skin = name // skins are PER PET — dressing the active one
		g.Save()
		label := "Classic"
		if name != "" {
			label = displayName(name)
		}
		p.flash = g.Pet.Name + " equipped " + label + "!"
		p.flashUntil = g.tick + 150
	}
	return nil
}

func (p *DressPage) Draw(g *Game, screen *ebiten.Image) {
	ui.DrawTextBold(screen, "Dress Up", 24, 28, 24, ui.Text)
	ui.DrawText(screen, "dressing "+g.Pet.Name+" — switch pets on Home", 24, 62, 12, ui.TextDim)

	// Live preview: the active pet exactly as it looks right now.
	g.DrawBlob(screen, ScreenW/2, 148)

	ui.DrawText(screen, "tap a look to equip it", 24, 208, 12, ui.TextDim)

	n := 1 + len(g.Sprites.SkinNames)
	for i := 0; i < n; i++ {
		name, _ := p.skinAt(g, i)
		x, y, w, h := p.tileRect(i)

		if g.Pet.Skin == name {
			ui.FillRoundRect(screen, float32(x-3), float32(y-3), float32(w+6), float32(h+6), 14, ui.Accent)
		}
		ui.FillRoundRect(screen, float32(x), float32(y), float32(w), float32(h), 12, ui.PanelHi)

		img := g.Sprites.Blob
		if name != "" {
			img = g.Sprites.Skins[name]
		}
		ui.DrawImageFit(screen, img, x+6, y+6, w-12, h-12)
	}

	if g.tick < p.flashUntil {
		ui.DrawTextCenter(screen, p.flash, ScreenW/2, 214, 12, ui.Gold, true)
	}
}

// displayName prettifies a skin file name ("natal" -> "Natal").
func displayName(name string) string {
	if name == "" {
		return "Classic"
	}
	return strings.ToUpper(name[:1]) + name[1:]
}
