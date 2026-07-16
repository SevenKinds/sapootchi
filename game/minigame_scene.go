package game

import (
	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/minigame"
	"sapootchi/ui"
)

// MiniGameScene runs any minigame.Game and applies its Result to the pet when
// done, then returns Home. This is the single seam through which every
// mini-game plugs in.
type MiniGameScene struct {
	game     minigame.Game
	applied  bool
	doneTick int
	summary  string
}

func NewMiniGameScene(mg minigame.Game) *MiniGameScene {
	return &MiniGameScene{game: mg}
}

func (s *MiniGameScene) Update(g *Game) error {
	if !s.game.Done() {
		return s.game.Update()
	}
	if !s.applied {
		s.apply(g)
		s.applied = true
		s.doneTick = g.tick
	}
	// Show the result briefly, then pop back Home.
	if g.tick-s.doneTick > 120 {
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

	got := 0
	for kind, n := range r.Items {
		p.AddFood(kind, n)
		got += n
	}

	// Build a human summary of what was earned.
	s.summary = "Score " + ui.Itoa(r.Score)
	if got > 0 {
		s.summary += "  ·  +" + ui.Itoa(got) + " food"
	}
	if r.Coins > 0 {
		s.summary += "  ·  +" + ui.Itoa(r.Coins) + " coins"
	}
	if r.StatDelta.Energy < 0 {
		s.summary += "  ·  energy burned"
	}
	g.Save()
}

func (s *MiniGameScene) Draw(g *Game, screen *ebiten.Image) {
	if s.game.Done() {
		ui.DrawTextCenter(screen, "Nice!", ScreenW/2, 260, 28, ui.Text, true)
		ui.DrawTextCenter(screen, s.summary, ScreenW/2, 300, 14, ui.Gold, true)
		return
	}
	s.game.Draw(screen)
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
