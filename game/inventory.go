package game

import (
	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/simulation"
	"sapootchi/ui"
)

// inventoryOrder fixes the display order of food kinds.
var inventoryOrder = []simulation.FoodKind{
	simulation.FoodApple,
	simulation.FoodSandwich,
	simulation.FoodCake,
}

// InventoryPage lists owned food; tapping Feed consumes one item.
type InventoryPage struct {
	flash      string
	flashUntil int
}

func (p *InventoryPage) Icon() ui.Icon { return ui.IconBag }
func (p *InventoryPage) Label() string { return "Items" }

func (p *InventoryPage) Update(g *Game) error {
	for i, kind := range inventoryOrder {
		if g.Pet.FoodCount(kind) <= 0 || !p.feedButton(i).Clicked() {
			continue
		}
		switch err := g.Pet.Feed(kind); err {
		case nil:
			p.flash = "Fed " + simulation.Foods[kind].Name + "!"
			g.Save()
		case simulation.ErrAsleep:
			p.flash = "Shh — the pet is asleep."
		case simulation.ErrPetAway:
			p.flash = "The pet ran away — check Home."
		default:
			p.flash = err.Error()
		}
		p.flashUntil = g.tick + 150
	}
	return nil
}

func (p *InventoryPage) Draw(g *Game, screen *ebiten.Image) {
	ui.DrawTextBold(screen, "Inventory", 24, 28, 24, ui.Text)
	ui.DrawText(screen, "Food from mini-games and the shop.", 24, 62, 13, ui.TextDim)

	empty := true
	for i, kind := range inventoryOrder {
		count := g.Pet.FoodCount(kind)
		def := simulation.Foods[kind]
		y := p.rowY(i)

		// Row panel.
		ui.FillRoundRect(screen, 24, float32(y), ScreenW-48, 72, 14, ui.Panel)

		// Count badge.
		ui.FillRoundRect(screen, 40, float32(y+20), 44, 32, 10, colIf(count > 0, ui.PanelHi, ui.Track))
		ui.DrawTextCenter(screen, "x"+ui.Itoa(count), 62, y+27, 14, colIf(count > 0, ui.Text, ui.TextDim), true)

		// Name + effect.
		ui.DrawTextBold(screen, def.Name, 100, y+14, 16, colIf(count > 0, ui.Text, ui.TextDim))
		effect := "+" + ui.Itoa(int(def.Hunger)) + "% hunger"
		if def.Happiness > 0 {
			effect += " · +happy"
		}
		ui.DrawText(screen, effect, 100, y+38, 12, ui.TextDim)

		// Feed action.
		if count > 0 {
			empty = false
			p.feedButton(i).Draw(screen, g.Pet.Awake())
		}
	}

	if empty {
		ui.DrawTextCenter(screen, "Nothing here yet — play Catch Food!", ScreenW/2,
			p.rowY(len(inventoryOrder))+24, 13, ui.TextDim, false)
	}

	if g.tick < p.flashUntil {
		ui.DrawTextCenter(screen, p.flash, ScreenW/2, 520, 13, ui.Gold, true)
	}
}

func (p *InventoryPage) rowY(i int) float64 { return float64(96 + i*88) }

func (p *InventoryPage) feedButton(i int) ui.Button {
	return ui.Button{X: ScreenW - 48 - 76, Y: p.rowY(i) + 16, W: 76, H: 40, Label: "Feed"}
}
