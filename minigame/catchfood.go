package minigame

import (
	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/simulation"
	"sapootchi/ui"
)

// Catch falling food -> feeds Hunger via food items. [SHIP IN POC]
//
// Move the basket with the mouse or arrow keys; catch food before it hits the
// floor. Final score awards 1-3 food items to inventory. This is feeding, so it
// does NOT spend energy — the Runner is the energy-burner.

const (
	catchDurationTicks = 25 * 60 // ~25s at 60 TPS
	basketW            = 60.0
	basketH            = 16.0
	itemR              = 11.0
	spawnEveryTicks    = 45
	fallSpeed          = 2.6
)

type fallingItem struct {
	x, y float64
	dead bool
}

// CatchFood is the POC mini-game.
type CatchFood struct {
	// Sprite, when set, draws the pet carrying the basket (the "real pet in
	// mini-games" setting).
	Sprite *ebiten.Image

	w, h    float64
	basketX float64
	items   []fallingItem
	score   int
	ticks   int
	done    bool
}

// NewCatchFood creates the game sized to the given play area.
func NewCatchFood(width, height int) *CatchFood {
	return &CatchFood{
		w:       float64(width),
		h:       float64(height),
		basketX: float64(width) / 2,
	}
}

func (c *CatchFood) Name() string { return "Catch Food" }

func (c *CatchFood) Update() error {
	if c.done {
		return nil
	}
	c.ticks++

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

	// Spawn items on a fixed cadence. Deterministic x from tick count keeps the
	// mini-game reproducible and avoids importing rand here.
	if c.ticks%spawnEveryTicks == 0 {
		x := float64((c.ticks*97)%int(c.w-2*itemR)) + itemR
		c.items = append(c.items, fallingItem{x: x, y: -itemR})
	}

	basketTop := c.h - 44
	for i := range c.items {
		it := &c.items[i]
		if it.dead {
			continue
		}
		it.y += fallSpeed
		if it.y >= basketTop-itemR && it.y <= basketTop+basketH &&
			it.x >= c.basketX-basketW/2 && it.x <= c.basketX+basketW/2 {
			it.dead = true
			c.score++
		} else if it.y > c.h {
			it.dead = true
		}
	}

	if c.ticks >= catchDurationTicks {
		c.done = true
	}
	return nil
}

func (c *CatchFood) Draw(screen *ebiten.Image) {
	for _, it := range c.items {
		if it.dead {
			continue
		}
		ui.FillRoundRect(screen, float32(it.x-itemR), float32(it.y-itemR),
			itemR*2, itemR*2, itemR, ui.Warn)
	}
	basketTop := c.h - 44
	// The pet peeks out from behind the basket when the real sprite is on.
	if c.Sprite != nil {
		ui.DrawImageFit(screen, c.Sprite, c.basketX-27, basketTop-36, 54, 38)
	}
	ui.FillRoundRect(screen, float32(c.basketX-basketW/2), float32(basketTop),
		basketW, basketH, 6, ui.Secondary)

	ui.DrawTextBold(screen, "CATCH FOOD — mouse / arrows", 12, 12, 13, ui.Text)
	ui.DrawText(screen, "caught: "+ui.Itoa(c.score), 12, 34, 14, ui.TextDim)
	secs := (catchDurationTicks - c.ticks) / 60
	ui.DrawText(screen, "time: "+ui.Itoa(secs)+"s", 12, 52, 14, ui.TextDim)
}

func (c *CatchFood) Done() bool { return c.done }

// Result maps score -> 1-3 food items (apples for the POC). Pays no coins, no
// energy (feeding is not exercise).
func (c *CatchFood) Result() Result {
	items := 1
	switch {
	case c.score >= 10:
		items = 3
	case c.score >= 5:
		items = 2
	}
	return Result{
		Score: c.score,
		Items: map[simulation.FoodKind]int{simulation.FoodApple: items},
	}
}

func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
