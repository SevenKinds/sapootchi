package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/minigame"
	"sapootchi/ui"
)

// gameEntry registers one mini-game in the Games page. Adding a game = add a
// line here (plus its implementation of minigame.Game). The factory receives
// the Game so entries can honor settings (e.g. the real-sprite toggle).
type gameEntry struct {
	name   string
	desc   string
	accent color.RGBA // badge color on the Games page (matches the stat it feeds)
	make   func(g *Game) minigame.Game
}

var gameCatalog = []gameEntry{
	{"Catch Food", "Catch treats — fills Hunger", ui.Warn, func(g *Game) minigame.Game {
		c := minigame.NewCatchFood(ScreenW, ScreenH)
		c.Sprite = g.minigameSprite()
		return c
	}},
	{"Runner", "Dodge & jump — burns Energy, pays coins", ui.Energy, func(g *Game) minigame.Game {
		r := minigame.NewRunner(ScreenW, ScreenH)
		r.Sprite = g.minigameSprite()
		return r
	}},
	{"Scrub", "Bath time — rub the dirt off, fills Hygiene", ui.Accent, func(g *Game) minigame.Game {
		s := minigame.NewScrub(ScreenW, ScreenH)
		s.Sprite = g.minigameSprite()
		return s
	}},
	{"Simon", "Repeat the sequence — fills Happiness", ui.Good, func(g *Game) minigame.Game {
		s := minigame.NewSimon(ScreenW, ScreenH)
		s.Sprite = g.minigameSprite()
		return s
	}},
	{"Arrows", "Hit the beat — timing game, wins Coffee", ui.Secondary, func(g *Game) minigame.Game {
		a := minigame.NewArrows(ScreenW, ScreenH)
		a.Sprite = g.minigameSprite()
		return a
	}},
	{"River", "Swim upstream — dodge rocks, grab coins", ui.Gold, func(g *Game) minigame.Game {
		r := minigame.NewRiver(ScreenW, ScreenH)
		r.Sprite = g.minigameSprite()
		return r
	}},
}

// minigameSprite returns the pet's current look (equipped skin or classic blob)
// when the settings toggle is on, or nil for the shape stand-ins.
func (g *Game) minigameSprite() *ebiten.Image {
	if g.Settings.RealSpriteInGames {
		return g.baseSprite()
	}
	return nil
}

// GamesPage lists the mini-games; picking one pushes it modally over the pager.
type GamesPage struct{}

func (p *GamesPage) Icon() ui.Icon { return ui.IconGames }
func (p *GamesPage) Label() string { return "Play" }

func (p *GamesPage) Update(g *Game) error {
	if !g.Pet.Awake() {
		return nil
	}
	for i, e := range gameCatalog {
		if p.rowButton(i).Clicked() {
			g.Push(NewMiniGameScene(e.make(g)))
		}
	}
	return nil
}

func (p *GamesPage) Draw(g *Game, screen *ebiten.Image) {
	ui.DrawTextBold(screen, "Mini-games", 24, 28, 24, ui.Text)
	ui.DrawText(screen, "Playing is how you care for your pet.", 24, 62, 13, ui.TextDim)

	awake := g.Pet.Awake()
	for i, e := range gameCatalog {
		b := p.rowButton(i)
		b.Draw(screen, awake)
		// Accent badge keyed to the stat the game feeds.
		badge := e.accent
		if !awake {
			badge = ui.Disabled
		}
		ui.FillRoundRect(screen, float32(b.X+12), float32(b.Y+12), 8, float32(b.H-24), 4, badge)
		ui.DrawTextBold(screen, e.name, b.X+32, b.Y+11, 16, colIf(awake, ui.Text, ui.TextDim))
		ui.DrawText(screen, e.desc, b.X+32, b.Y+36, 11, colIf(awake, ui.Text, ui.TextDim))
	}

	if !awake {
		msg := "The pet is asleep — come back when it wakes."
		if g.Pet.Away {
			msg = "The pet ran away — leave food out on Home."
		}
		ui.DrawTextCenter(screen, msg, ScreenW/2, 90+float64(len(gameCatalog))*76+12, 12, ui.TextDim, false)
	}
}

func (p *GamesPage) rowButton(i int) ui.Button {
	return ui.Button{X: 24, Y: float64(90 + i*76), W: ScreenW - 48, H: 64, Secondary: true}
}
