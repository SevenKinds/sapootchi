package game

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/simulation"
	"sapootchi/ui"
)

// HomePage is the main tab: greet, watch, and care for the pet.
type HomePage struct {
	flash      string
	flashUntil int
	greeted    bool
}

func (p *HomePage) Icon() ui.Icon { return ui.IconHome }
func (p *HomePage) Label() string { return "Home" }

func (p *HomePage) setFlash(g *Game, msg string) {
	p.flash = msg
	p.flashUntil = g.tick + 150 // ~2.5s
}

func (p *HomePage) Update(g *Game) error {
	// The rescue moment: a brand-new pet was found in bad shape.
	if !p.greeted {
		p.greeted = true
		if g.Pet.Age(time.Now()) < 2*time.Minute {
			p.flash = "You found him just in time — take care of him!"
			p.flashUntil = g.tick + 420 // ~7s
		}
	}

	// Tap the header to switch between pets (when there is more than one).
	if len(g.Pets) > 1 && ui.Tapped(16, 16, ScreenW-32, 60) {
		pet := g.SwitchPet()
		p.setFlash(g, "Now caring for "+pet.Name)
	}

	if g.Pet.Away {
		if !g.Pet.FoodLeftOut && p.leaveFoodButton().Clicked() {
			g.Pet.LeaveFoodOut()
			g.Save()
		}
		return nil
	}

	// Tap the pet -> a brand reaction animation plays in a modal. (Modal, not
	// in-place: the animations show the classic SAPO, whatever skin is worn.)
	if ui.Tapped(ScreenW/2-105, 85, 210, 175) {
		if frames, ok := g.Sprites.Anims[p.pickAnim(g)]; ok {
			g.Push(NewAnimModal(frames, g.current()))
			return nil
		}
	}

	if !g.Pet.Awake() {
		return nil
	}
	for _, b := range p.careButtons() {
		if !b.Clicked() {
			continue
		}
		switch b.Label {
		case "Feed":
			p.feedBest(g)
		case "Bathe":
			g.Pet.Bathe()
		case "Rest":
			g.Pet.Rest()
		case "Pet":
			g.Pet.Pet()
		}
		p.checkPerfectCare(g)
		g.Save()
	}
	return nil
}

// pickAnim chooses a reaction that fits the pet's state right now.
func (p *HomePage) pickAnim(g *Game) string {
	pet := g.Pet
	switch {
	case pet.Asleep:
		return "sleepy"
	case pet.Mood() == simulation.MoodHungry || pet.Mood() == simulation.MoodLonely:
		return "triste"
	case pet.Energized():
		return "estrelas"
	default:
		return []string{"wink", "coracoes", "estrelas"}[g.Rng.Intn(3)]
	}
}

func (p *HomePage) feedBest(g *Game) {
	for _, k := range []simulation.FoodKind{simulation.FoodApple, simulation.FoodSandwich, simulation.FoodCake} {
		if g.Pet.FoodCount(k) > 0 {
			_ = g.Pet.Feed(k)
			p.setFlash(g, "Fed "+simulation.Foods[k].Name)
			return
		}
	}
	p.setFlash(g, "No food! Play Catch Food.")
}

func (p *HomePage) checkPerfectCare(g *Game) {
	if !g.Pet.PerfectCare() {
		return
	}
	yd := time.Now().YearDay()
	if g.lastPerfectYearDay == yd {
		return
	}
	g.lastPerfectYearDay = yd
	g.Pet.Coins += 10
	p.setFlash(g, "Perfect Care! +10 coins")
}

func (p *HomePage) Draw(g *Game, screen *ebiten.Image) {
	pet := g.Pet

	p.drawHeader(g, screen)

	if pet.Away {
		p.drawAway(g, screen)
		return
	}

	awake := pet.Awake()

	// Pet + state pill (shared view — same look on the Inventory page).
	drawPetAndState(g, screen, 185, 262)

	// Stats panel.
	ui.FillRoundRect(screen, 16, 292, ScreenW-32, 150, 14, ui.Panel)
	const bx, bw = 36, ScreenW - 72
	ui.StatBar(screen, "Happiness", pet.Stats.Happiness, bx, 322, bw, nil)
	ui.StatBar(screen, "Hunger", pet.Stats.Hunger, bx, 356, bw, nil)
	ui.StatBar(screen, "Hygiene", pet.Stats.Hygiene, bx, 390, bw, nil)
	ui.StatBar(screen, "Energy", pet.Stats.Energy, bx, 424, bw, ui.Energy)

	for _, b := range p.careButtons() {
		b.Draw(screen, awake)
	}

	if g.tick < p.flashUntil {
		ui.DrawTextCenter(screen, p.flash, ScreenW/2, 536, 13, ui.Gold, true)
	}
}

func (p *HomePage) drawHeader(g *Game, screen *ebiten.Image) {
	pet := g.Pet
	ui.FillRoundRect(screen, 16, 16, ScreenW-32, 60, 12, ui.Panel)
	name := pet.Name
	if len(g.Pets) > 1 {
		name += "  " + ui.Itoa(g.Active+1) + "/" + ui.Itoa(len(g.Pets))
	}
	ui.DrawTextBold(screen, name, 32, 26, 18, ui.Text)
	sub := pet.Personality.String() + " · " + pet.Phase.String()
	if len(g.Pets) > 1 {
		sub += " · tap to switch"
	}
	ui.DrawText(screen, sub, 32, 50, 12, ui.TextDim)

	// Coins, right-aligned with a gold dot.
	coinStr := ui.Itoa(pet.Coins)
	cw := ui.TextWidth(coinStr, 16, true)
	rightX := float64(ScreenW - 32)
	ui.DrawTextBold(screen, coinStr, rightX-cw, 30, 16, ui.Gold)
	ui.FillCircle(screen, float32(rightX-cw-14), 38.5, 6.5, ui.Gold)
}

func (p *HomePage) drawAway(g *Game, screen *ebiten.Image) {
	ui.DrawTextCenter(screen, g.Pet.Name+" ran away!", ScreenW/2, 210, 22, ui.Text, true)
	if g.Pet.FoodLeftOut {
		ui.DrawTextCenter(screen, "Food is out. He might wander back...", ScreenW/2, 250, 13, ui.TextDim, false)
		ui.DrawTextCenter(screen, "(28% chance each day)", ScreenW/2, 272, 12, ui.TextDim, false)
	} else {
		ui.DrawTextCenter(screen, "His hunger hit zero.", ScreenW/2, 250, 13, ui.TextDim, false)
		p.leaveFoodButton().Draw(screen, true)
	}
}

func (p *HomePage) leaveFoodButton() ui.Button {
	return ui.Button{X: (ScreenW - 200) / 2, Y: 320, W: 200, H: 46, Label: "Leave food out"}
}

func (p *HomePage) careButtons() []ui.Button {
	const y, h, m, gp = 470.0, 48.0, 16.0, 8.0
	w := (ScreenW - 2*m - 3*gp) / 4
	labels := []string{"Feed", "Bathe", "Rest", "Pet"}
	out := make([]ui.Button, 4)
	for i, l := range labels {
		out[i] = ui.Button{X: m + float64(i)*(w+gp), Y: y, W: w, H: h, Label: l}
	}
	return out
}
