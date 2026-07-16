package game

import (
	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/simulation"
	"sapootchi/ui"
)

// shopItem is a coin sink. POC: buy food; cosmetics/decor come later.
type shopItem struct {
	label string
	price int
	buy   func(p *simulation.Pet)
}

// ShopPage is a minimal coin sink so earning coins feels good in the POC.
type ShopPage struct {
	flash      string
	flashUntil int
}

func (p *ShopPage) Icon() ui.Icon { return ui.IconShop }
func (p *ShopPage) Label() string { return "Shop" }

func (p *ShopPage) items() []shopItem {
	return []shopItem{
		{"Apple  ·  +25% hunger", 10, func(pet *simulation.Pet) { pet.AddFood(simulation.FoodApple, 1) }},
		{"Sandwich  ·  +50% hunger", 18, func(pet *simulation.Pet) { pet.AddFood(simulation.FoodSandwich, 1) }},
		{"Cake  ·  +happiness", 25, func(pet *simulation.Pet) { pet.AddFood(simulation.FoodCake, 1) }},
	}
}

func (p *ShopPage) Update(g *Game) error {
	for i, it := range p.items() {
		if !p.rowButton(i).Clicked() {
			continue
		}
		if g.Pet.Coins >= it.price {
			g.Pet.Coins -= it.price
			it.buy(g.Pet)
			p.flash = "Bought " + it.label
			g.Save()
		} else {
			p.flash = "Not enough coins"
		}
		p.flashUntil = g.tick + 150
	}
	return nil
}

func (p *ShopPage) Draw(g *Game, screen *ebiten.Image) {
	ui.DrawTextBold(screen, "Shop", 24, 28, 24, ui.Text)

	coinStr := ui.Itoa(g.Pet.Coins)
	cw := ui.TextWidth(coinStr, 16, true)
	rightX := float64(ScreenW - 24)
	ui.DrawTextBold(screen, coinStr, rightX-cw, 32, 16, ui.Gold)
	ui.FillCircle(screen, float32(rightX-cw-14), 40.5, 6.5, ui.Gold)

	for i, it := range p.items() {
		b := p.rowButton(i)
		affordable := g.Pet.Coins >= it.price
		b.Draw(screen, affordable)
		ui.DrawTextBold(screen, it.label, b.X+18, b.Y+14, 15, colIf(affordable, ui.Text, ui.TextDim))
		ui.DrawText(screen, ui.Itoa(it.price)+" coins", b.X+18, b.Y+36, 12, colIf(affordable, ui.Gold, ui.TextDim))
	}

	if g.tick < p.flashUntil {
		ui.DrawTextCenter(screen, p.flash, ScreenW/2, 500, 13, ui.Gold, true)
	}
}

func (p *ShopPage) rowButton(i int) ui.Button {
	return ui.Button{X: 24, Y: float64(84 + i*84), W: ScreenW - 48, H: 68, Secondary: true}
}
