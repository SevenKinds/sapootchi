package game

import (
	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/minigame"
	"sapootchi/simulation"
	"sapootchi/ui"
)

// MiniGameScene runs any minigame.Game and applies its Result to the pet when
// done, then shows a results card until the player taps. This is the single
// seam through which every mini-game plugs in.
type MiniGameScene struct {
	game     minigame.Game
	applied  bool
	doneTick int
	rewards  []string
}

func NewMiniGameScene(mg minigame.Game) *MiniGameScene {
	return &MiniGameScene{game: mg}
}

// quitRect is the leave (X) button shown during play, top-right corner.
func quitRect() (x, y, w, h float64) { return ScreenW - 40, 8, 32, 32 }

func (s *MiniGameScene) Update(g *Game) error {
	if !s.game.Done() {
		// Leave mid-game: back to the pager, no rewards.
		if ui.Tapped(quitRect()) {
			g.Pop()
			return nil
		}
		return s.game.Update()
	}
	if !s.applied {
		s.apply(g)
		s.applied = true
		s.doneTick = g.tick
	}
	// Tap to dismiss (small grace period so a final game tap doesn't skip it).
	if g.tick-s.doneTick > 20 && ui.Tapped(0, 0, ScreenW, ScreenH) {
		g.Pop()
	}
	return nil
}

func (s *MiniGameScene) apply(g *Game) {
	r := s.game.Result()
	p := g.Pet
	p.Coins += r.Coins
	p.Stats.Happiness = clamp01(p.Stats.Happiness + r.StatDelta.Happiness)
	p.Stats.Hunger = clamp01(p.Stats.Hunger + r.StatDelta.Hunger)
	p.Stats.Hygiene = clamp01(p.Stats.Hygiene + r.StatDelta.Hygiene)
	p.Stats.Energy = clamp01(p.Stats.Energy + r.StatDelta.Energy)
	p.Hidden.Intelligence = clamp01(p.Hidden.Intelligence + r.Hidden.Intelligence)
	p.Hidden.Friendship = clamp01(p.Hidden.Friendship + r.Hidden.Friendship)

	// Itemized reward lines for the card (fixed kind order — maps iterate randomly).
	s.rewards = s.rewards[:0]
	for _, kind := range inventoryOrder {
		if n := r.Items[kind]; n > 0 {
			p.AddFood(kind, n)
			s.rewards = append(s.rewards, "+"+ui.Itoa(n)+" "+simulation.Foods[kind].Name)
		}
	}
	if r.Coins > 0 {
		s.rewards = append(s.rewards, "+"+ui.Itoa(r.Coins)+" coins")
	}
	for _, e := range []struct {
		v    float64
		name string
	}{
		{r.StatDelta.Happiness, "happiness"},
		{r.StatDelta.Hunger, "hunger"},
		{r.StatDelta.Hygiene, "hygiene"},
		{r.StatDelta.Energy, "energy"},
	} {
		if e.v > 0 {
			s.rewards = append(s.rewards, "+"+ui.Itoa(int(e.v))+"% "+e.name)
		}
	}
	if r.StatDelta.Energy < 0 {
		s.rewards = append(s.rewards, ui.Itoa(int(r.StatDelta.Energy))+"% energy — good exercise!")
	}
	if r.Hidden.Intelligence > 0 {
		s.rewards = append(s.rewards, "it feels a little smarter...")
	}
	g.Save()
}

func (s *MiniGameScene) Draw(g *Game, screen *ebiten.Image) {
	if !s.game.Done() {
		s.game.Draw(screen)
		// Leave button on top of every game.
		qx, qy, qw, qh := quitRect()
		cx, cy := float32(qx+qw/2), float32(qy+qh/2)
		ui.FillCircle(screen, cx, cy, 14, ui.Panel)
		ui.StrokeLine(screen, cx-5, cy-5, cx+5, cy+5, 2.5, ui.TextDim)
		ui.StrokeLine(screen, cx-5, cy+5, cx+5, cy-5, 2.5, ui.TextDim)
		return
	}

	// Results card.
	cardH := 150.0 + float64(len(s.rewards))*22
	y := (ScreenH - cardH) / 2
	ui.FillRoundRect(screen, 36, float32(y+3), ScreenW-72, float32(cardH), 18, ui.Shadow)
	ui.FillRoundRect(screen, 36, float32(y), ScreenW-72, float32(cardH), 18, ui.Panel)

	ui.DrawTextCenter(screen, s.game.Name(), ScreenW/2, y+22, 13, ui.TextDim, true)
	ui.DrawTextCenter(screen, "Score "+ui.Itoa(s.game.Result().Score), ScreenW/2, y+44, 26, ui.Text, true)

	ly := y + 92
	for _, line := range s.rewards {
		ui.DrawTextCenter(screen, line, ScreenW/2, ly, 13, ui.Gold, false)
		ly += 22
	}

	ui.DrawTextCenter(screen, "tap to continue", ScreenW/2, y+cardH-26, 11, ui.TextDim, false)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}
