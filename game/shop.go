package game

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/simulation"
	"sapootchi/ui"
)

// shopItem is a coin sink. available==false rows render disabled with a reason.
type shopItem struct {
	label     string
	price     int
	available bool
	reason    string               // shown when unavailable
	buy       func(g *Game) string // returns flash message
}

// ShopPage sells food, coffee — and the tadpole, which hatches a whole new pet.
type ShopPage struct {
	flash      string
	flashUntil int
}

func (p *ShopPage) Icon() ui.Icon { return ui.IconShop }
func (p *ShopPage) Label() string { return "Shop" }

func foodBuy(kind simulation.FoodKind, label string) func(g *Game) string {
	return func(g *Game) string {
		g.Pet.AddFood(kind, 1)
		return "Bought " + label
	}
}

func (p *ShopPage) items(g *Game) []shopItem {
	roster := len(g.Pets)
	return []shopItem{
		{"Apple  ·  +25% hunger", 10, true, "", foodBuy(simulation.FoodApple, "Apple")},
		{"Steak  ·  +50% hunger", 18, true, "", foodBuy(simulation.FoodSandwich, "Steak")},
		{"Berries  ·  +happiness", 25, true, "", foodBuy(simulation.FoodCake, "Berries")},
		{"Coffee  ·  +35% energy", 20, true, "", foodBuy(simulation.FoodCoffee, "Coffee")},
		{"Tadpole  ·  a whole new pet!", 100, roster < MaxPets, "pet limit reached",
			func(g *Game) string {
				pet := hatchTadpole(g)
				return pet.Name + " hatched! (" + pet.Personality.String() + ")"
			}},
	}
}

// hatchTadpole creates the new pet: healthy-ish (it was adopted, not rescued)
// and with a NATURE DIFFERENT from every pet you already have.
func hatchTadpole(g *Game) *simulation.Pet {
	taken := map[simulation.Personality]bool{}
	for _, p := range g.Pets {
		taken[p.Personality] = true
	}
	// Collect free personalities; if all are taken, any random one.
	var free []simulation.Personality
	for pp := simulation.PersonalityLazy; pp <= simulation.PersonalityCurious; pp++ {
		if !taken[pp] {
			free = append(free, pp)
		}
	}
	personality := simulation.Personality(g.Rng.Intn(4))
	if len(free) > 0 {
		personality = free[g.Rng.Intn(len(free))]
	}

	names := []string{"Blobby", "Girino", "Bolhas"}
	name := names[min(len(g.Pets), len(names)-1)]

	pet := simulation.NewPet(name, personality, time.Now())
	// Adopted, not rescued: hatches in decent shape.
	pet.Stats = simulation.Stats{Happiness: 60, Hunger: 60, Hygiene: 60, Energy: 60}
	g.AddPet(pet)
	return pet
}

// Skin unlocks: this many on display at once; buying one reveals the next.
const (
	skinPrice   = 150
	skinsOnSale = 3
)

func (p *ShopPage) Update(g *Game) error {
	for i, it := range p.items(g) {
		if !p.rowButton(i).Clicked() {
			continue
		}
		switch {
		case !it.available:
			p.flash = it.reason
		case g.Pet.Coins >= it.price:
			g.Pet.Coins -= it.price
			p.flash = it.buy(g)
			g.Save()
		default:
			p.flash = "Not enough coins"
		}
		p.flashUntil = g.tick + 150
	}

	// Skin tiles: confirm before spending.
	for i, name := range p.skinsForSale(g) {
		if !ui.Tapped(p.skinTile(i)) {
			continue
		}
		if g.Pet.Coins < skinPrice {
			p.flash = "Not enough coins"
			p.flashUntil = g.tick + 150
			continue
		}
		skin := name // capture
		g.Push(NewConfirmModal(
			"Unlock "+displayName(skin)+"?",
			"a new look for any of your pets",
			g.Sprites.Skins[skin], skinPrice, g.current(),
			func(g *Game) {
				g.Pet.Coins -= skinPrice
				g.Settings.OwnedSkins = append(g.Settings.OwnedSkins, skin)
				p.flash = displayName(skin) + " unlocked — equip it in Dress!"
				p.flashUntil = g.tick + 150
				g.Save()
			}))
	}
	return nil
}

// skinsForSale: the next few locked looks; buying reveals the one behind.
func (p *ShopPage) skinsForSale(g *Game) []string {
	un := g.UnownedSkins()
	if len(un) > skinsOnSale {
		un = un[:skinsOnSale]
	}
	return un
}

func (p *ShopPage) Draw(g *Game, screen *ebiten.Image) {
	ui.DrawTextBold(screen, "Shop", 24, 28, 24, ui.Text)

	coinStr := ui.Itoa(g.Pet.Coins)
	cw := ui.TextWidth(coinStr, 16, true)
	rightX := float64(ScreenW - 24)
	ui.DrawTextBold(screen, coinStr, rightX-cw, 32, 16, ui.Gold)
	g.DrawCoin(screen, rightX-cw-24, 32, 18)

	for i, it := range p.items(g) {
		b := p.rowButton(i)
		enabled := it.available && g.Pet.Coins >= it.price
		b.Draw(screen, enabled)
		ui.DrawTextBold(screen, it.label, b.X+16, b.Y+8, 14, colIf(enabled, ui.Text, ui.TextDim))
		sub := ui.Itoa(it.price) + " coins"
		if !it.available {
			sub = it.reason
		}
		ui.DrawText(screen, sub, b.X+16, b.Y+29, 11, colIf(enabled, ui.Gold, ui.TextDim))
	}

	// Looks on sale.
	sale := p.skinsForSale(g)
	if len(sale) > 0 {
		ui.DrawTextBold(screen, "Looks · "+ui.Itoa(skinPrice)+" coins each", 24, 424, 13, ui.TextDim)
		ui.DrawText(screen, "buying one reveals the next", 24, 444, 10, ui.TextDim)
		for i, name := range sale {
			x, y, w, h := p.skinTile(i)
			affordable := g.Pet.Coins >= skinPrice
			ui.FillRoundRect(screen, float32(x), float32(y), float32(w), float32(h), 12,
				colIf(affordable, ui.PanelHi, ui.Panel))
			ui.DrawImageFit(screen, g.Sprites.Skins[name], x+8, y+6, w-16, h-30)
			ui.DrawTextCenter(screen, displayName(name), x+w/2, y+h-20, 10, colIf(affordable, ui.Text, ui.TextDim), true)
		}
	}

	if g.tick < p.flashUntil {
		ui.DrawTextCenter(screen, p.flash, ScreenW/2, 560, 13, ui.Gold, true)
	}
}

func (p *ShopPage) rowButton(i int) ui.Button {
	return ui.Button{X: 24, Y: float64(70 + i*66), W: ScreenW - 48, H: 56, Secondary: true}
}

func (p *ShopPage) skinTile(i int) (x, y, w, h float64) {
	const tw, gap = 96.0, 10.0
	total := 3*tw + 2*gap
	x0 := (ScreenW - total) / 2
	return x0 + float64(i)*(tw+gap), 456, tw, 92
}
