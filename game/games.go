package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/minigame"
	"sapootchi/simulation"
	"sapootchi/ui"
)

// gameEntry registers one mini-game in the Games page. Adding a game = add a
// line here (plus its implementation of minigame.Game). The factory receives
// the Game so entries can honor settings (e.g. the real-sprite toggle).
// earns/earnsCoins render as icons on the card: what playing can pay out.
type gameEntry struct {
	name       string
	glyph      rune       // card icon (Nerd Font)
	accent     color.RGBA // matches the stat the game feeds
	earns      []simulation.FoodKind
	earnsCoins bool
	make       func(g *Game) minigame.Game
}

// Scrub is launched from Home's Bathe button (bath time IS the mini-game), so
// it is not listed in the catalog.
func newScrubGame(g *Game) minigame.Game {
	s := minigame.NewScrub(ScreenW, ScreenH)
	s.Sprite = g.minigameSprite()
	return s
}

var gameCatalog = []gameEntry{
	{"Catch Food", '\U000F0DF1', ui.Warn,
		[]simulation.FoodKind{simulation.FoodApple, simulation.FoodSandwich, simulation.FoodCake}, false,
		func(g *Game) minigame.Game {
			c := minigame.NewCatchFood(ScreenW, ScreenH)
			c.Fruits = g.Sprites.Fruits
			c.Meat = g.Sprites.Meat
			return c
		}},
	{"Arrows", '\uf001', ui.Secondary,
		[]simulation.FoodKind{simulation.FoodCoffee}, false,
		func(g *Game) minigame.Game {
			a := minigame.NewArrows(ScreenW, ScreenH)
			a.Sprite = g.minigameSprite()
			return a
		}},
	{"Climber", '\uf148', ui.Energy,
		nil, true,
		func(g *Game) minigame.Game {
			c := minigame.NewClimber(ScreenW, ScreenH)
			c.Sprite = g.minigameSprite()
			c.Platforms = g.Sprites.Platforms
			c.Clouds = g.Sprites.Clouds
			c.CoinFrames = g.Sprites.Coin
			return c
		}},
	{"Simon", '\uf0eb', ui.Good,
		nil, true,
		func(g *Game) minigame.Game {
			s := minigame.NewSimon(ScreenW, ScreenH)
			s.Sprite = g.minigameSprite()
			return s
		}},
	{"River", '\uf043', ui.Gold,
		nil, true,
		func(g *Game) minigame.Game {
			r := minigame.NewRiver(ScreenW, ScreenH)
			r.Sprite = g.minigameSprite()
			r.Rocks = g.Sprites.RiverRocks
			r.Foam = g.Sprites.Foam
			r.CoinFrames = g.Sprites.Coin
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

// GamesPage is a card grid: each mini-game is a tile with its icon, name, and
// what it pays. Tap to play.
type GamesPage struct{}

func (p *GamesPage) Icon() ui.Icon { return ui.IconGames }
func (p *GamesPage) Label() string { return "Play" }

// Card grid geometry.
const (
	gameCols    = 2
	gameCardW   = 152.0
	gameCardH   = 112.0
	gameCardGap = 12.0
	gameGridY   = 96.0
)

func (p *GamesPage) cardRect(i int) (x, y, w, h float64) {
	total := gameCols*gameCardW + (gameCols-1)*gameCardGap
	x0 := (ScreenW - total) / 2
	col, row := i%gameCols, i/gameCols
	return x0 + float64(col)*(gameCardW+gameCardGap), gameGridY + float64(row)*(gameCardH+gameCardGap), gameCardW, gameCardH
}

func (p *GamesPage) Update(g *Game) error {
	if !g.Pet.Awake() {
		return nil
	}
	for i, e := range gameCatalog {
		if ui.Tapped(p.cardRect(i)) {
			g.Push(NewMiniGameScene(e.make(g)))
		}
	}
	return nil
}

func (p *GamesPage) Draw(g *Game, screen *ebiten.Image) {
	ui.DrawTextBold(screen, "Play", 24, 28, 24, ui.Text)
	ui.DrawText(screen, "playing is how you care for your pet", 24, 62, 13, ui.TextDim)

	awake := g.Pet.Awake()
	cx, cy := ui.Cursor()

	for i, e := range gameCatalog {
		x, y, w, h := p.cardRect(i)

		bg := ui.Panel
		if awake {
			bg = ui.PanelHi
			if cx >= x && cx <= x+w && cy >= y && cy <= y+h {
				bg = ui.Disabled // hover lift
			}
		}
		ui.FillRoundRect(screen, float32(x), float32(y+2), float32(w), float32(h), 16, ui.Shadow)
		ui.FillRoundRect(screen, float32(x), float32(y), float32(w), float32(h), 16, bg)

		// Accent keel along the top.
		accent := e.accent
		if !awake {
			accent = ui.Disabled
		}
		ui.FillRoundRect(screen, float32(x+16), float32(y), float32(w-32), 4, 2, accent)

		ui.DrawGlyph(screen, e.glyph, x+w/2, y+36, 26, accent)
		ui.DrawTextCenter(screen, e.name, x+w/2, y+56, 14, colIf(awake, ui.Text, ui.TextDim), true)

		// Payout row, centered: item icons and/or the spinning coin.
		n := len(e.earns)
		if e.earnsCoins {
			n++
		}
		ix := x + w/2 - float64(n-1)*13
		for _, kind := range e.earns {
			drawItemIcon(screen, kind, ix, y+90, 15)
			ix += 26
		}
		if e.earnsCoins {
			g.DrawCoin(screen, ix-8, y+90-8, 16)
		}
	}

	if !awake {
		msg := "the pet is asleep — come back when it wakes"
		if g.Pet.Away {
			msg = "the pet ran away — leave food out on Home"
		}
		rows := (len(gameCatalog) + gameCols - 1) / gameCols
		ui.DrawTextCenter(screen, msg, ScreenW/2, gameGridY+float64(rows)*(gameCardH+gameCardGap)+10, 12, ui.TextDim, false)
	}
}
