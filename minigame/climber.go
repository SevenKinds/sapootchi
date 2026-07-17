package minigame

import (
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/simulation"
	"sapootchi/ui"
)

// Climber is the energy-burner: a vertical platform climber. The pet starts on
// the ground; your first touch sets him bouncing NON-STOP, and you steer him
// left/right (hold a screen side, or arrows/A-D) from platform to platform.
// Fall off the bottom and the climb is over. Score = height.
//
// Platform types (brackeys sheet): green = solid, blue = drifts sideways,
// orange = crumbles after one bounce.

const (
	climbGravity = 0.42
	climbJumpV   = -11.5
	climbSteer   = 0.55 // horizontal acceleration while steering
	climbMaxVX   = 4.6
	climbPetSize = 40.0
	climbPlatH   = 12.0
	climbTimeCap = 120 * 60 // safety cap; falling is the real end

	climbEnergyCost = 30.0
)

type platKind int

const (
	platSolid platKind = iota
	platMoving
	platCrumble
)

type platform struct {
	x, y   float64 // center x, top y
	w      float64
	kind   platKind
	vx     float64
	broken bool
}

type climbCoin struct {
	x, y float64
	got  bool
}

// Climber implements minigame.Game.
type Climber struct {
	// Injected art (the minigame package owns no assets).
	Sprite     *ebiten.Image   // the pet, when the real-sprite setting is on
	Platforms  *ebiten.Image   // brackeys platform sheet (64x64)
	Clouds     []*ebiten.Image // background clouds
	CoinFrames []*ebiten.Image // the game's spinning coin (pickups)

	w, h    float64
	rng     *rand.Rand
	started bool

	petX, petY float64 // world coords (y grows down; 0 = ground)
	vx, vy     float64

	plats   []platform
	coins   []climbCoin
	topY    float64 // highest platform generated so far (world y)
	camY    float64 // world y at the TOP of the screen
	height  float64 // best height climbed (px)
	pickups int
	ticks   int
	done    bool
}

// NewClimber creates the game sized to the play area.
func NewClimber(width, height int) *Climber {
	c := &Climber{
		w:   float64(width),
		h:   float64(height),
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	c.petX = c.w / 2
	c.petY = -climbPetSize / 2 // standing on the ground (y=0)
	c.camY = -c.h
	// Ground is a full-width solid platform at y=0.
	c.plats = append(c.plats, platform{x: c.w / 2, y: 0, w: c.w, kind: platSolid})
	c.topY = 0
	c.genUpTo(-c.h * 2)
	return c
}

func (c *Climber) Name() string { return "Climber" }

// genUpTo generates platforms upward (decreasing world y) until limit.
func (c *Climber) genUpTo(limit float64) {
	for c.topY > limit {
		gap := 58 + c.rng.Float64()*34
		// Higher up, gaps stretch a little.
		gap += math.Min(30, -c.topY/220)
		y := c.topY - gap

		w := 52 + c.rng.Float64()*26
		p := platform{
			x: w/2 + 8 + c.rng.Float64()*(c.w-w-16),
			y: y, w: w,
		}
		switch r := c.rng.Float64(); {
		case r < 0.16 && y < -600:
			p.kind = platCrumble
		case r < 0.38 && y < -300:
			p.kind = platMoving
			p.vx = 0.7 + c.rng.Float64()*0.9
			if c.rng.Float64() < 0.5 {
				p.vx = -p.vx
			}
		default:
			p.kind = platSolid
		}
		c.plats = append(c.plats, p)

		if c.rng.Float64() < 0.28 {
			c.coins = append(c.coins, climbCoin{x: p.x, y: y - 26})
		}
		c.topY = y
	}
}

func (c *Climber) Update() error {
	if c.done {
		return nil
	}
	c.ticks++

	if !c.started {
		// Waiting at ground level: the first touch/press starts the climb.
		if ui.PointerJustPressed() ||
			ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyUp) {
			c.started = true
			c.vy = climbJumpV
		}
		return nil
	}

	c.steer()

	// Physics.
	c.petX += c.vx
	c.petY += c.vy
	c.vy += climbGravity

	// Wrap horizontally (classic climber feel).
	if c.petX < -climbPetSize/2 {
		c.petX = c.w + climbPetSize/2
	} else if c.petX > c.w+climbPetSize/2 {
		c.petX = -climbPetSize / 2
	}

	// Bounce on platforms — only while falling.
	if c.vy > 0 {
		bottom := c.petY + climbPetSize/2
		for i := range c.plats {
			p := &c.plats[i]
			if p.broken {
				continue
			}
			if bottom >= p.y && bottom <= p.y+climbPlatH+c.vy &&
				c.petX > p.x-p.w/2-climbPetSize*0.3 && c.petX < p.x+p.w/2+climbPetSize*0.3 {
				c.vy = climbJumpV
				if p.kind == platCrumble {
					p.broken = true
				}
				break
			}
		}
	}

	// Moving platforms drift and rebound off the banks.
	for i := range c.plats {
		p := &c.plats[i]
		if p.kind != platMoving {
			continue
		}
		p.x += p.vx
		if p.x-p.w/2 < 4 || p.x+p.w/2 > c.w-4 {
			p.vx = -p.vx
		}
	}

	// Coins.
	for i := range c.coins {
		co := &c.coins[i]
		if !co.got && math.Hypot(c.petX-co.x, c.petY-co.y) < climbPetSize*0.7 {
			co.got = true
			c.pickups++
		}
	}

	// Camera follows the pet upward; height is the best climb.
	if c.petY < c.camY+c.h*0.42 {
		c.camY = c.petY - c.h*0.42
	}
	if h := -c.petY; h > c.height {
		c.height = h
	}
	c.genUpTo(c.camY - c.h)

	// Cull far-below platforms/coins.
	if len(c.plats) > 60 {
		alive := c.plats[:0]
		for _, p := range c.plats {
			if p.y < c.camY+c.h*2 {
				alive = append(alive, p)
			}
		}
		c.plats = alive
	}

	// Fell off the bottom -> done.
	if c.petY > c.camY+c.h+climbPetSize {
		c.done = true
	}
	if c.ticks >= climbTimeCap {
		c.done = true
	}
	return nil
}

// steer accelerates toward held input: left/right halves of the screen, or keys.
func (c *Climber) steer() {
	dir := 0.0
	if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		dir = -1
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		dir = 1
	}
	if dir == 0 && ui.PointerHeld() {
		if px, _ := ui.Cursor(); px < c.w/2 {
			dir = -1
		} else {
			dir = 1
		}
	}
	if dir == 0 {
		c.vx *= 0.93 // glide to a stop
		return
	}
	c.vx = clampF(c.vx+dir*climbSteer, -climbMaxVX, climbMaxVX)
}

// platform sheet slices (brackeys 64x64): rows green/tan/orange/blue, wide
// platform at (16..48, row*16, w32 h10).
func (c *Climber) platSprite(kind platKind) *ebiten.Image {
	if c.Platforms == nil {
		return nil
	}
	row := 0 // green: solid
	switch kind {
	case platMoving:
		row = 3 // blue
	case platCrumble:
		row = 2 // orange
	}
	return c.Platforms.SubImage(imageRect(16, row*16, 48, row*16+10)).(*ebiten.Image)
}

func (c *Climber) Draw(screen *ebiten.Image) {
	// Background clouds, parallax on camera height.
	for i, img := range c.Clouds {
		if i >= 4 {
			break
		}
		scale := 0.30 + float64(i)*0.06
		w := float64(img.Bounds().Dx()) * scale
		span := c.w + w
		x := math.Mod(float64(i)*160-c.camY*0.05*(1+float64(i)*0.4), span)
		if x < 0 {
			x += span
		}
		y := 80 + float64(i)*130 + math.Mod(-c.camY*0.12, c.h)
		if y > c.h {
			y -= c.h + 100
		}
		ui.DrawImageNearest(screen, img, x-w, y, scale, 0.14)
	}

	// Platforms.
	for _, p := range c.plats {
		if p.broken {
			continue
		}
		sy := p.y - c.camY
		if sy < -20 || sy > c.h+20 {
			continue
		}
		if spr := c.platSprite(p.kind); spr != nil {
			sw := float64(spr.Bounds().Dx())
			ui.DrawImageNearest(screen, spr, p.x-p.w/2, sy, p.w/sw, 1)
		} else {
			ui.FillRoundRect(screen, float32(p.x-p.w/2), float32(sy), float32(p.w), climbPlatH, 5, ui.Good)
		}
	}

	// Coins.
	for _, co := range c.coins {
		if co.got {
			continue
		}
		sy := co.y - c.camY
		if sy > -12 && sy < c.h+12 {
			drawSpinCoin(screen, c.CoinFrames, c.ticks, co.x, sy, 20)
		}
	}

	c.drawPet(screen)

	// HUD.
	ui.DrawTextBold(screen, "CLIMBER", 14, 14, 15, ui.Text)
	ui.DrawText(screen, "hold a side to steer — don't fall!", 14, 34, 11, ui.TextDim)
	hStr := ui.Itoa(int(c.height/10)) + "m  ·  coins " + ui.Itoa(c.pickups)
	ui.DrawTextBold(screen, hStr, c.w-52-ui.TextWidth(hStr, 13, true), 14, 13, ui.Gold)

	if !c.started {
		ui.DrawTextCenter(screen, "Tap to start climbing!", c.w/2, c.h/2-60, 20, ui.Text, true)
	}
}

func (c *Climber) drawPet(screen *ebiten.Image) {
	sy := clampF(1.0-c.vy*0.012, 0.80, 1.22)
	sx := 2 - sy
	w := climbPetSize * sx
	hh := climbPetSize * sy
	x := c.petX - w/2
	y := (c.petY - c.camY) - hh/2

	if c.Sprite != nil {
		// Real art keeps its aspect — squash-stretch distortion reads badly on
		// the detailed sprites (only the shape stand-in stretches).
		ui.DrawImageFit(screen, c.Sprite, c.petX-climbPetSize/2, (c.petY-c.camY)-climbPetSize/2,
			climbPetSize, climbPetSize)
		return
	}
	ui.FillRoundRect(screen, float32(x), float32(y), float32(w), float32(hh), 12, ui.Good)
	for _, ex := range []float64{x + w*0.42, x + w*0.72} {
		ui.FillRoundRect(screen, float32(ex-6), float32(y+hh*0.34-6), 12, 12, 6, color.White)
		ui.FillRoundRect(screen, float32(ex-2), float32(y+hh*0.34-2), 4, 4, 2, color.RGBA{0x10, 0x14, 0x1a, 0xff})
	}
}

func (c *Climber) Done() bool { return c.done }

// Result: pays for height and pickups; the big energy burn.
func (c *Climber) Result() Result {
	return Result{
		Score:     int(c.height / 10), // meters
		Coins:     int(c.height/100) + c.pickups*2,
		StatDelta: simulation.Stats{Energy: -climbEnergyCost, Happiness: 8},
	}
}
