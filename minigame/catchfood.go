package minigame

import (
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/simulation"
	"sapootchi/ui"
)

// Catch falling food -> feeds Hunger via food items. Move the basket; catch the
// good stuff, dodge the rotten ones. What you catch is what you keep: caught
// food banks into inventory by kind.
//
// Uses real randomness (like Simon): per-run variety IS the game — a fixed
// pattern made every round identical.

const (
	catchDurationTicks = 30 * 60
	basketW            = 60.0
	basketH            = 16.0
	itemR              = 11.0
)

type catchKind int

const (
	catchApple catchKind = iota
	catchSandwich
	catchCake
	catchRotten
)

type fallingItem struct {
	kind catchKind
	x, y float64
	vy   float64
	dead bool
}

// CatchFood is the feeding mini-game.
type CatchFood struct {
	// Sprite, when set, draws the pet carrying the basket.
	Sprite *ebiten.Image

	w, h      float64
	rng       *rand.Rand
	basketX   float64
	items     []fallingItem
	caught    [4]int // per kind
	score     int
	nextSpawn int
	burstLeft int
	splatT    int // rotten-catch feedback timer
	ticks     int
	done      bool
}

// NewCatchFood creates the game sized to the given play area.
func NewCatchFood(width, height int) *CatchFood {
	return &CatchFood{
		w:       float64(width),
		h:       float64(height),
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
		basketX: float64(width) / 2,
	}
}

func (c *CatchFood) Name() string { return "Catch Food" }

// difficulty ramps: spawns get denser, falls get faster.
func (c *CatchFood) spawnGap() int {
	g := 50 - c.ticks/120
	if g < 22 {
		g = 22
	}
	return g
}

func (c *CatchFood) fallSpeed() float64 {
	s := 2.4 + float64(c.ticks)*0.0011
	if s > 4.8 {
		s = 4.8
	}
	return s
}

func (c *CatchFood) rollKind() catchKind {
	r := c.rng.Float64()
	switch {
	case r < 0.60:
		return catchApple
	case r < 0.72:
		return catchSandwich
	case r < 0.82:
		return catchCake
	default:
		return catchRotten
	}
}

func (c *CatchFood) Update() error {
	if c.done {
		return nil
	}
	c.ticks++
	if c.splatT > 0 {
		c.splatT--
	}

	// Basket control: pointer follows (mouse or touch), arrows nudge.
	if mx, _ := ui.Cursor(); mx > 0 {
		c.basketX = mx
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		c.basketX -= 5
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		c.basketX += 5
	}
	c.basketX = clampF(c.basketX, basketW/2, c.w-basketW/2)

	// Spawning: random cadence with occasional 3-item bursts.
	c.nextSpawn--
	if c.nextSpawn <= 0 {
		c.spawn()
		if c.burstLeft > 0 {
			c.burstLeft--
			c.nextSpawn = 10
		} else {
			c.nextSpawn = c.spawnGap() + c.rng.Intn(18)
			if c.rng.Float64() < 0.14 {
				c.burstLeft = 2
			}
		}
	}

	basketTop := c.h - 44
	for i := range c.items {
		it := &c.items[i]
		if it.dead {
			continue
		}
		it.y += it.vy
		if it.y >= basketTop-itemR && it.y <= basketTop+basketH &&
			it.x >= c.basketX-basketW/2 && it.x <= c.basketX+basketW/2 {
			it.dead = true
			c.catch(it.kind)
		} else if it.y > c.h {
			it.dead = true
		}
	}

	if c.ticks >= catchDurationTicks {
		c.done = true
	}
	return nil
}

func (c *CatchFood) spawn() {
	c.items = append(c.items, fallingItem{
		kind: c.rollKind(),
		x:    itemR + c.rng.Float64()*(c.w-2*itemR),
		y:    -itemR,
		vy:   c.fallSpeed() * (0.9 + c.rng.Float64()*0.35),
	})
}

func (c *CatchFood) catch(k catchKind) {
	c.caught[k]++
	switch k {
	case catchApple:
		c.score++
	case catchSandwich, catchCake:
		c.score += 2
	case catchRotten:
		c.score -= 2
		if c.score < 0 {
			c.score = 0
		}
		c.splatT = 30
	}
}

var catchColors = [4]color.RGBA{
	{0xe6, 0x56, 0x4a, 0xff}, // apple: red
	{0xe8, 0xb8, 0x6d, 0xff}, // sandwich: bread
	{0xf7, 0xa8, 0xc4, 0xff}, // cake: pink
	{0x6b, 0x72, 0x3a, 0xff}, // rotten: murky
}

func (c *CatchFood) Draw(screen *ebiten.Image) {
	for _, it := range c.items {
		if it.dead {
			continue
		}
		clr := catchColors[it.kind]
		ui.FillCircle(screen, float32(it.x), float32(it.y), itemR, clr)
		if it.kind == catchRotten { // X marks the bad one
			ui.StrokeLine(screen, float32(it.x-5), float32(it.y-5), float32(it.x+5), float32(it.y+5), 2, ui.Track)
			ui.StrokeLine(screen, float32(it.x-5), float32(it.y+5), float32(it.x+5), float32(it.y-5), 2, ui.Track)
		}
	}

	basketTop := c.h - 44
	if c.Sprite != nil {
		ui.DrawImageFit(screen, c.Sprite, c.basketX-27, basketTop-36, 54, 38)
	}
	basketClr := ui.Secondary
	if c.splatT > 0 {
		basketClr = ui.Bad
	}
	ui.FillRoundRect(screen, float32(c.basketX-basketW/2), float32(basketTop),
		basketW, basketH, 6, basketClr)

	ui.DrawTextBold(screen, "CATCH FOOD", 14, 14, 15, ui.Text)
	ui.DrawText(screen, "catch the good, dodge the rotten", 14, 34, 11, ui.TextDim)
	scoreStr := "score " + ui.Itoa(c.score)
	ui.DrawTextBold(screen, scoreStr, c.w-52-ui.TextWidth(scoreStr, 15, true), 14, 15, ui.Gold)
	secs := (catchDurationTicks - c.ticks) / 60
	tStr := ui.Itoa(secs) + "s"
	ui.DrawText(screen, tStr, c.w-52-ui.TextWidth(tStr, 12, false), 34, 12, ui.TextDim)
	if c.splatT > 0 {
		ui.DrawTextCenter(screen, "yuck!", c.basketX, basketTop-56, 14, ui.Bad, true)
	}
}

func (c *CatchFood) Done() bool { return c.done }

// Result: what you caught is what you keep — every 2 catches of a kind banks
// one item of that kind (at least 1 apple if you scored at all). No coins; no
// energy (feeding is not exercise).
func (c *CatchFood) Result() Result {
	items := map[simulation.FoodKind]int{}
	bank := func(kind simulation.FoodKind, caught int) {
		if n := caught / 2; n > 0 {
			items[kind] = n
		}
	}
	bank(simulation.FoodApple, c.caught[catchApple])
	bank(simulation.FoodSandwich, c.caught[catchSandwich])
	bank(simulation.FoodCake, c.caught[catchCake])
	if len(items) == 0 && c.score > 0 {
		items[simulation.FoodApple] = 1
	}
	return Result{Score: c.score, Items: items}
}
