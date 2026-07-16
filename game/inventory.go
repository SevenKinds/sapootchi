package game

import (
	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/simulation"
	"sapootchi/ui"
)

// inventoryOrder fixes the display order of item slots.
var inventoryOrder = []simulation.FoodKind{
	simulation.FoodApple,
	simulation.FoodSandwich,
	simulation.FoodCake,
	simulation.FoodCoffee,
}

// invOrder is the slots actually shown: the base items, plus dev-only items
// when in dev mode or when some are held (a release build must still show
// pills spawned in a dev session).
func invOrder(g *Game) []simulation.FoodKind {
	order := inventoryOrder
	if DevMode || g.Pet.FoodCount(simulation.FoodEnergyPill) > 0 {
		order = append(append([]simulation.FoodKind{}, order...), simulation.FoodEnergyPill)
	}
	return order
}

// InventoryPage shows Blobby (in whatever state he's in — same as Home) and a
// shelf of item icons. Two ways to give him an item: DRAG it onto him, or TAP
// it to select and then tap him (friendlier on touch). If he's asleep, he
// can't take it — same rule as everywhere else.
type InventoryPage struct {
	dragging   bool
	dragKind   simulation.FoodKind
	selected   bool
	selKind    simulation.FoodKind
	flash      string
	flashUntil int
}

func (p *InventoryPage) Icon() ui.Icon { return ui.IconBag }
func (p *InventoryPage) Label() string { return "Items" }

// Layout (design-space).
const (
	invBlobCy  = 200.0
	invPillY   = 300.0
	invShelfY  = 448.0
	invSlotW   = 66.0
	invSlotGap = 12.0
)

func (p *InventoryPage) slotRect(g *Game, i int) (x, y, w, h float64) {
	n := float64(len(invOrder(g)))
	// Shrink slots when there are more of them (dev items add a 5th).
	slotW := invSlotW
	if maxW := (ScreenW - 40 - (n-1)*invSlotGap) / n; maxW < slotW {
		slotW = maxW
	}
	total := n*slotW + (n-1)*invSlotGap
	x0 := (ScreenW - total) / 2
	return x0 + float64(i)*(slotW+invSlotGap), invShelfY, slotW, slotW
}

// blobZone is the drop target around the pet.
func (p *InventoryPage) blobZone() (x, y, w, h float64) {
	return ScreenW/2 - 110, invBlobCy - 110, 220, 220
}

// CapturesPress claims presses on stocked slots so the pager doesn't treat an
// item drag as a page swipe.
func (p *InventoryPage) CapturesPress(g *Game, x, y float64) bool {
	if g.Pet.Away {
		return false
	}
	for i, kind := range invOrder(g) {
		if g.Pet.FoodCount(kind) <= 0 {
			continue
		}
		sx, sy, sw, sh := p.slotRect(g, i)
		if x >= sx && x <= sx+sw && y >= sy && y <= sy+sh {
			return true
		}
	}
	return false
}

func (p *InventoryPage) setFlash(g *Game, msg string) {
	p.flash = msg
	p.flashUntil = g.tick + 150
}

func (p *InventoryPage) Update(g *Game) error {
	// Start a drag from a stocked slot.
	if !p.dragging && ui.PointerJustPressed() && !g.Pet.Away {
		px, py := ui.PressPos()
		for i, kind := range invOrder(g) {
			if g.Pet.FoodCount(kind) <= 0 {
				continue
			}
			sx, sy, sw, sh := p.slotRect(g, i)
			if px >= sx && px <= sx+sw && py >= sy && py <= sy+sh {
				p.dragging = true
				p.dragKind = kind
				break
			}
		}
	}

	// Release: either a drop on Blobby (drag-feed) or a tap (select-feed).
	if p.dragging && ui.PointerJustReleased() {
		p.dragging = false
		cx, cy := ui.Cursor()
		bx, by, bw, bh := p.blobZone()
		if cx >= bx && cx <= bx+bw && cy >= by && cy <= by+bh {
			p.give(g, p.dragKind)
			return nil
		}
		// Not a drop — was it a tap on a slot? Toggle selection.
		for i, kind := range invOrder(g) {
			if g.Pet.FoodCount(kind) > 0 && ui.Tapped(p.slotRect(g, i)) {
				if p.selected && p.selKind == kind {
					p.selected = false
				} else {
					p.selected, p.selKind = true, kind
				}
			}
		}
		return nil
	}

	// Tap Blobby with an item selected -> give it.
	if p.selected && !g.Pet.Away && ui.Tapped(p.blobZone()) {
		p.give(g, p.selKind)
		if g.Pet.FoodCount(p.selKind) <= 0 {
			p.selected = false // stack ran out
		}
	}
	return nil
}

func (p *InventoryPage) give(g *Game, kind simulation.FoodKind) {
	def := simulation.Foods[kind]
	switch err := g.Pet.Feed(kind); err {
	case nil:
		verb := map[string]string{"": "ate", "Drink": "drank", "Use": "used"}[def.Verb]
		p.setFlash(g, g.Pet.Name+" "+verb+" the "+def.Name+"!")
		g.Save()
	case simulation.ErrAsleep:
		p.setFlash(g, "Zzz... he's asleep — he can't eat now.")
	case simulation.ErrPetAway:
		p.setFlash(g, "He ran away — check Home.")
	default:
		p.setFlash(g, err.Error())
	}
}

func (p *InventoryPage) Draw(g *Game, screen *ebiten.Image) {
	ui.DrawTextBold(screen, "Inventory", 24, 28, 24, ui.Text)

	// Blobby, in the same state as on Home.
	home := drawPetAndState(g, screen, invBlobCy, invPillY)

	// Shelf.
	total := 0
	for i, kind := range invOrder(g) {
		count := g.Pet.FoodCount(kind)
		total += count
		sx, sy, sw, sh := p.slotRect(g, i)

		bg := ui.Panel
		if count > 0 {
			bg = ui.PanelHi
		}
		// Selection ring.
		if p.selected && p.selKind == kind {
			ui.FillRoundRect(screen, float32(sx-3), float32(sy-3), float32(sw+6), float32(sh+6), 16, ui.Accent)
		}
		ui.FillRoundRect(screen, float32(sx), float32(sy), float32(sw), float32(sh), 14, bg)

		if count > 0 {
			// Hide the shelf copy of the icon while it's being dragged... only
			// if it's the last one; otherwise the stack remains.
			if !(p.dragging && p.dragKind == kind && count == 1) {
				drawItemIcon(screen, kind, sx+sw/2, sy+sh/2)
			}
			// Count badge.
			ui.FillCircle(screen, float32(sx+sw-10), float32(sy+12), 10, ui.Accent)
			ui.DrawTextCenter(screen, ui.Itoa(count), sx+sw-10, sy+5, 11, ui.Text, true)
		}
	}

	// Hints.
	switch {
	case total == 0:
		ui.DrawTextCenter(screen, "Nothing here yet — play Catch Food!", ScreenW/2, invShelfY+invSlotW+18, 13, ui.TextDim, false)
	case p.dragging:
		def := simulation.Foods[p.dragKind]
		ui.DrawTextCenter(screen, def.Name+" · "+effectText(def), ScreenW/2, invShelfY-28, 12, ui.Text, true)
	case p.selected:
		def := simulation.Foods[p.selKind]
		ui.DrawTextCenter(screen, "tap "+g.Pet.Name+" to give the "+def.Name+" · "+effectText(def), ScreenW/2, invShelfY-28, 12, ui.Text, true)
	case home:
		ui.DrawTextCenter(screen, "drag an item to "+g.Pet.Name+" — or tap to select", ScreenW/2, invShelfY-28, 12, ui.TextDim, false)
	}

	// The dragged item follows the pointer.
	if p.dragging {
		cx, cy := ui.Cursor()
		drawItemIcon(screen, p.dragKind, cx, cy-8)
	}

	if g.tick < p.flashUntil {
		ui.DrawTextCenter(screen, p.flash, ScreenW/2, invShelfY+invSlotW+18, 13, ui.Gold, true)
	}
}

// effectText describes what consuming an item does, built from its definition.
func effectText(def simulation.FoodDef) string {
	s := ""
	if def.Hunger > 0 {
		s += "+" + ui.Itoa(int(def.Hunger)) + "% hunger"
	}
	if def.Energy > 0 {
		if s != "" {
			s += " · "
		}
		s += "+" + ui.Itoa(int(def.Energy)) + "% energy"
	}
	if def.Happiness > 0 {
		if s != "" {
			s += " · "
		}
		s += "+happy"
	}
	return s
}
