package game

import (
	"image/color"
	"math"
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
	hearts     []heartFx
}

// heartFx is a floating heart from petting — drawn OVER the pet, so it works
// on every skin (the hearts-pose swap only exists for the classic look).
type heartFx struct {
	x, y, vx, vy float64
	life, max    int
}

func (p *HomePage) Icon() ui.Icon { return ui.IconHome }
func (p *HomePage) Label() string { return "Home" }

func (p *HomePage) setFlash(g *Game, msg string) {
	p.flash = msg
	p.flashUntil = g.tick + 150 // ~2.5s
}

// Home layout: the pet is the hero, stats ride the right edge.
const (
	homeBlobCx = ScreenW / 2.0
	homeBlobCy = 262.0
	homePillY  = 392.0
)

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

	// The evolution moment: the sim says it's time -> the modal does the rest.
	if g.Pet.CanEvolve(time.Now()) {
		g.Push(NewEvolveModal(g.current()))
		return nil
	}

	// Tap the pet -> a brand reaction animation plays in a modal. (Modal, not
	// in-place: the animations show the classic SAPO, whatever skin is worn.)
	if ui.Tapped(homeBlobCx-105, homeBlobCy-110, 210, 220) {
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
			// Shortcut to the Items page — feeding is inventory-driven.
			g.Main.GoTo(PageItems)
		case "Bathe":
			// Bath time IS the Scrub mini-game.
			g.Push(NewMiniGameScene(newScrubGame(g)))
		case "Rest":
			// Tuck in: voluntary nap until FULL energy — confirmed first,
			// since he's unavailable while he sleeps.
			g.Push(NewConfirmAction(
				"Tuck "+g.Pet.Name+" in?",
				"he'll sleep until energy is back to 100%",
				g.Sprites.Asleep, "Tuck in", g.current(),
				func(g *Game) {
					g.Pet.Rest()
					p.setFlash(g, "Sweet dreams — back at 100%")
					g.Save()
				}))
		case "Pet":
			// Affection moment: hearts always (particles work on any skin;
			// the pose swap plays too on the classic look). The happiness
			// bonus is rate-limited so spamming is love, not progress.
			if g.Pet.Pet(time.Now()) {
				g.ShowReaction("hearts", 100)
				p.spawnHearts(g, 6)
				p.setFlash(g, "+happiness")
			} else {
				g.ShowReaction("wink", 70)
				p.spawnHearts(g, 2)
			}
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

	// Clouds drift way in the back, before anything else.
	p.drawClouds(g, screen)

	p.drawHeader(g, screen)

	if pet.Away {
		p.drawAway(g, screen)
		return
	}

	awake := pet.Awake()

	// Pet front and center — the stats live above their matching care button.
	drawPetAndState(g, screen, homeBlobCx, homeBlobCy, homePillY)
	p.updateAndDrawHearts(screen)
	p.drawStatsAboveButtons(g, screen)

	for _, b := range p.careButtons() {
		b.Draw(screen, awake)
	}

	if g.tick < p.flashUntil {
		ui.DrawTextCenter(screen, p.flash, ScreenW/2, 442, 13, ui.Gold, true)
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

	// Coins, right-aligned with the spinning coin.
	coinStr := ui.Itoa(pet.Coins)
	cw := ui.TextWidth(coinStr, 16, true)
	rightX := float64(ScreenW - 32)
	ui.DrawTextBold(screen, coinStr, rightX-cw, 30, 16, ui.Gold)
	g.DrawCoin(screen, rightX-cw-24, 30, 18)
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

// Care row: quiet icon buttons (Nerd Font glyphs), sitting just above the tab
// bar.
func (p *HomePage) careButtons() []ui.GlyphButton {
	const y, h, m, gp = 514.0, 52.0, 20.0, 10.0
	defs := []struct {
		label string
		glyph rune
		clr   color.RGBA
	}{
		{"Feed", '\uf0f5', ui.Warn},      // cutlery -> hunger
		{"Bathe", '\uf043', ui.Energy},   // droplet -> hygiene
		{"Pet", '\uf004', ui.Bad},        // heart -> happiness
		{"Rest", '\uf186', ui.Secondary}, // moon -> energy
	}
	w := (ScreenW - 2*m - 3*gp) / 4
	out := make([]ui.GlyphButton, len(defs))
	for i, d := range defs {
		out[i] = ui.GlyphButton{
			X: m + float64(i)*(w+gp), Y: y, W: w, H: h,
			Glyph: d.glyph, GlyphColor: d.clr, Label: d.label,
		}
	}
	return out
}

// drawStatsAboveButtons puts each stat right above the button that feeds it:
// hunger/Feed, hygiene/Bathe, happiness/Pet, energy/Rest. The bar becomes the
// button's gauge.
func (p *HomePage) drawStatsAboveButtons(g *Game, screen *ebiten.Image) {
	// statCells order: happiness, hunger, hygiene, energy -> button columns.
	order := []int{1, 2, 0, 3}
	for col, b := range p.careButtons() {
		c := statCells[order[col]]
		v := c.get(g.Pet.Stats)
		clr := c.fixed
		if clr == nil {
			clr = statColorFor(v)
		}
		y := b.Y - 44
		ui.DrawGlyph(screen, c.glyph, b.X+9, y+8, 12, clr)
		ui.DrawTextBold(screen, ui.Itoa(int(v+0.5))+"%", b.X+21, y, 11, ui.Text)
		ui.FillRoundRect(screen, float32(b.X), float32(y+19), float32(b.W), 6, 3, ui.Track)
		if fw := b.W * v / 100; fw > 6 {
			ui.FillRoundRect(screen, float32(b.X), float32(y+19), float32(fw), 6, 3, clr)
		}
	}
}

// spawnHearts scatters floating hearts around the pet.
func (p *HomePage) spawnHearts(g *Game, n int) {
	for i := 0; i < n; i++ {
		p.hearts = append(p.hearts, heartFx{
			x:    homeBlobCx + (g.Rng.Float64()*2-1)*70,
			y:    homeBlobCy + (g.Rng.Float64()*2-1)*40,
			vx:   (g.Rng.Float64()*2 - 1) * 0.4,
			vy:   -0.8 - g.Rng.Float64()*0.7,
			life: 70 + g.Rng.Intn(30),
		})
		p.hearts[len(p.hearts)-1].max = p.hearts[len(p.hearts)-1].life
	}
}

func (p *HomePage) updateAndDrawHearts(screen *ebiten.Image) {
	alive := p.hearts[:0]
	for _, h := range p.hearts {
		h.x += h.vx
		h.y += h.vy
		h.life--
		if h.life > 0 {
			alpha := uint8(255 * h.life / h.max)
			size := 10 + 4*float64(h.max-h.life)/float64(h.max)
			ui.DrawGlyph(screen, '\uf004', h.x, h.y, size, color.RGBA{0xe6, 0x56, 0x4a, alpha})
			alive = append(alive, h)
		}
	}
	p.hearts = alive
}

// drawClouds drifts the cloud sprites slowly across the far background.
func (p *HomePage) drawClouds(g *Game, screen *ebiten.Image) {
	if len(g.Sprites.Clouds) == 0 {
		return
	}
	layers := []struct {
		y, scale, speed float64
		alpha           float32
	}{
		{34, 0.34, 0.10, 0.16},
		{92, 0.26, 0.06, 0.12},
		{236, 0.40, 0.14, 0.18},
		{330, 0.22, 0.04, 0.10},
	}
	for i, l := range layers {
		img := g.Sprites.Clouds[(i*3)%len(g.Sprites.Clouds)]
		w := float64(img.Bounds().Dx()) * l.scale
		span := ScreenW + w
		x := ScreenW - math.Mod(float64(g.tick)*l.speed+float64(i)*210, span)
		ui.DrawImageNearest(screen, img, x-w, l.y, l.scale, l.alpha)
	}
}
