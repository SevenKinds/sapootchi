package minigame

import (
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"sapootchi/simulation"
	"sapootchi/ui"
)

// Runner is the endless-dodge mini-game and the DEDICATED energy-burner.
// Jump ground blocks — but DON'T jump into the flying bars — and grab the
// floating coins for bonus pay.
//
// Uses real randomness: obstacle patterns vary per run.

const (
	runnerDurationTicks = 40 * 60
	runnerStartDelay    = 48
	runnerGroundOffset  = 96
	runnerPetX          = 56
	runnerPetSize       = 42
	runnerObW           = 26
	runnerGravity       = 1.0
	runnerJumpV         = -15.0
	runnerBaseSpeed     = 4.2
	runnerSpeedRamp     = 0.0018
	runnerMaxSpeed      = 11.0

	runnerBarH = 14.0 // flying bar thickness

	runnerEnergyCost = 30.0
)

type obstacle struct {
	x      float64
	h      float64 // ground block height; for flying bars, altitude is fixed
	flying bool
	passed bool
}

type coinPickup struct {
	x, y float64
	got  bool
}

// Runner implements minigame.Game.
type Runner struct {
	// Sprite, when set, is drawn as the player character.
	Sprite *ebiten.Image

	w, h      float64
	rng       *rand.Rand
	petY      float64
	vy        float64
	onGround  bool
	obstacles []obstacle
	coins     []coinPickup
	pickups   int
	nextGap   int
	lastSpawn int
	scroll    float64
	score     int
	ticks     int
	done      bool
}

// NewRunner creates the game sized to the play area.
func NewRunner(width, height int) *Runner {
	r := &Runner{
		w:   float64(width),
		h:   float64(height),
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	r.petY = r.groundTop()
	r.onGround = true
	r.lastSpawn = -999
	r.nextGap = 90
	return r
}

func (r *Runner) groundY() float64   { return r.h - runnerGroundOffset }
func (r *Runner) groundTop() float64 { return r.groundY() - runnerPetSize }
func (r *Runner) running() bool      { return r.ticks >= runnerStartDelay }

func (r *Runner) speed() float64 {
	s := runnerBaseSpeed + float64(r.ticks-runnerStartDelay)*runnerSpeedRamp
	if s > runnerMaxSpeed {
		s = runnerMaxSpeed
	}
	return s
}

func (r *Runner) Name() string { return "Runner" }

func (r *Runner) Update() error {
	if r.done {
		return nil
	}
	r.ticks++

	if r.jumpPressed() && r.onGround {
		r.vy = runnerJumpV
	}
	r.petY += r.vy
	r.vy += runnerGravity
	if r.petY >= r.groundTop() {
		r.petY = r.groundTop()
		r.vy = 0
		r.onGround = true
	} else {
		r.onGround = false
	}

	if !r.running() {
		return nil
	}

	speed := r.speed()
	r.scroll += speed

	// Spawn with jittered gaps; mix ground blocks, flying bars, and coins.
	if r.ticks-r.lastSpawn >= r.nextGap {
		r.spawn()
		r.lastSpawn = r.ticks
		base := 88 - (r.ticks-runnerStartDelay)/120
		if base < 42 {
			base = 42
		}
		r.nextGap = base + r.rng.Intn(30)
	}

	// Move + judge obstacles.
	alive := r.obstacles[:0]
	for _, o := range r.obstacles {
		o.x -= speed
		if !o.passed && o.x+runnerObW < runnerPetX {
			o.passed = true
			r.score++
		}
		if r.collides(o) {
			r.done = true
		}
		if o.x+runnerObW > 0 {
			alive = append(alive, o)
		}
	}
	r.obstacles = alive

	// Coins.
	keep := r.coins[:0]
	for _, c := range r.coins {
		c.x -= speed
		if !c.got && r.overlapsCoin(c) {
			c.got = true
			r.pickups++
		}
		if !c.got && c.x > -12 {
			keep = append(keep, c)
		}
	}
	r.coins = keep

	if r.ticks >= runnerDurationTicks {
		r.done = true
	}
	return nil
}

func (r *Runner) spawn() {
	roll := r.rng.Float64()
	switch {
	case roll < 0.22 && r.ticks > runnerStartDelay+500:
		// Flying bar: stay grounded to pass under it.
		r.obstacles = append(r.obstacles, obstacle{x: r.w + runnerObW, flying: true})
	default:
		heights := []float64{30, 38, 46, 54}
		h := heights[r.rng.Intn(len(heights))]
		r.obstacles = append(r.obstacles, obstacle{x: r.w + runnerObW, h: h})
		// Risk/reward: sometimes a coin floats above the block.
		if r.rng.Float64() < 0.4 {
			r.coins = append(r.coins, coinPickup{x: r.w + runnerObW + 12, y: r.groundY() - h - 58})
		}
	}
	// Free-floating coin at grab height.
	if r.rng.Float64() < 0.25 {
		r.coins = append(r.coins, coinPickup{x: r.w + 150, y: r.groundY() - 24})
	}
}

// barTop/barBottom: the flying bar sits where a grounded pet fits under it.
func (r *Runner) barTop() float64 { return r.groundY() - 64 - runnerBarH }

func (r *Runner) collides(o obstacle) bool {
	if o.flying {
		return runnerPetX < o.x+runnerObW+14 &&
			runnerPetX+runnerPetSize > o.x &&
			r.petY < r.barTop()+runnerBarH
	}
	oTop := r.groundY() - o.h
	return runnerPetX < o.x+runnerObW &&
		runnerPetX+runnerPetSize > o.x &&
		r.petY+runnerPetSize > oTop
}

func (r *Runner) overlapsCoin(c coinPickup) bool {
	const cr = 10
	return c.x+cr > runnerPetX && c.x-cr < runnerPetX+runnerPetSize &&
		c.y+cr > r.petY && c.y-cr < r.petY+runnerPetSize
}

func (r *Runner) jumpPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsKeyJustPressed(ebiten.KeyUp) ||
		inpututil.IsKeyJustPressed(ebiten.KeyW) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
		len(inpututil.AppendJustPressedTouchIDs(nil)) > 0
}

func (r *Runner) Draw(screen *ebiten.Image) {
	r.drawParallax(screen)

	gy := r.groundY()
	ui.FillRoundRect(screen, 0, float32(gy), float32(r.w), float32(runnerGroundOffset), 0, ui.Track)
	ui.FillRoundRect(screen, 0, float32(gy), float32(r.w), 4, 2, ui.PanelHi)
	for x := -(int(r.scroll) % 34); x < int(r.w); x += 34 {
		ui.FillRoundRect(screen, float32(x), float32(gy+12), 16, 4, 2, ui.Panel)
	}

	for _, o := range r.obstacles {
		if o.flying {
			ui.FillRoundRect(screen, float32(o.x), float32(r.barTop()), runnerObW+14, runnerBarH, 6, ui.Secondary)
			ui.FillRoundRect(screen, float32(o.x), float32(r.barTop()), runnerObW+14, 4, 2,
				color.RGBA{0x95, 0x7a, 0xff, 0xff})
			continue
		}
		ui.FillRoundRect(screen, float32(o.x), float32(gy-o.h), runnerObW, float32(o.h), 5, ui.Bad)
		ui.FillRoundRect(screen, float32(o.x), float32(gy-o.h), runnerObW, 5, 2.5,
			color.RGBA{0xff, 0x82, 0x78, 0xff})
	}

	for _, c := range r.coins {
		if !c.got {
			ui.FillCircle(screen, float32(c.x), float32(c.y), 9, ui.Gold)
			ui.FillCircle(screen, float32(c.x), float32(c.y), 4.5, color.RGBA{0xc9, 0x9e, 0x1f, 0xff})
		}
	}

	r.drawPet(screen)

	ui.DrawTextBold(screen, "RUNNER", 14, 14, 15, ui.Text)
	ui.DrawText(screen, "jump blocks — duck under bars", 14, 34, 11, ui.TextDim)
	scoreStr := "cleared " + ui.Itoa(r.score) + "  ·  coins " + ui.Itoa(r.pickups)
	ui.DrawTextBold(screen, scoreStr, r.w-52-ui.TextWidth(scoreStr, 13, true), 14, 13, ui.Gold)
	secs := (runnerDurationTicks - r.ticks) / 60
	tStr := ui.Itoa(secs) + "s"
	ui.DrawText(screen, tStr, r.w-52-ui.TextWidth(tStr, 12, false), 34, 12, ui.TextDim)

	if !r.running() {
		ui.DrawTextCenter(screen, "Get ready!", r.w/2, r.h/2-40, 26, ui.Text, true)
	}
}

func (r *Runner) drawParallax(screen *ebiten.Image) {
	ys := []float64{120, 180, 250, 320, 400}
	for i, y := range ys {
		layer := 0.3 + float64(i)*0.15
		off := int(r.scroll * layer)
		for x := -(off % 90); x < int(r.w); x += 90 {
			ui.FillRoundRect(screen, float32(x), float32(y), 26, 3, 1.5,
				color.RGBA{0xff, 0xff, 0xff, 0x14})
		}
	}
}

func (r *Runner) drawPet(screen *ebiten.Image) {
	sy := 1.0
	if !r.onGround {
		sy = clampF(1.0-r.vy*0.010, 0.82, 1.20)
	}
	sx := 2 - sy
	w := runnerPetSize * sx
	hh := runnerPetSize * sy
	bx := runnerPetX + (runnerPetSize-w)/2
	by := (r.petY + runnerPetSize) - hh

	if r.Sprite != nil {
		ui.DrawImageStretch(screen, r.Sprite, bx, by, w, hh)
		return
	}

	ui.FillRoundRect(screen, float32(bx), float32(by), float32(w), float32(hh), 12, ui.Good)
	eyeR := 6.0
	ex1 := bx + w*0.42
	ex2 := bx + w*0.72
	ey := by + hh*0.34
	for _, ex := range []float64{ex1, ex2} {
		ui.FillRoundRect(screen, float32(ex-eyeR), float32(ey-eyeR), float32(eyeR*2), float32(eyeR*2), float32(eyeR), color.White)
		ui.FillRoundRect(screen, float32(ex-2), float32(ey-2), 4, 4, 2, color.RGBA{0x10, 0x14, 0x1a, 0xff})
	}
}

func (r *Runner) Done() bool { return r.done }

// Result: coins by distance + picked-up coins, a big energy burn, a little
// happiness from play.
func (r *Runner) Result() Result {
	return Result{
		Score:     r.score,
		Coins:     r.score*3 + r.pickups*2,
		StatDelta: simulation.Stats{Energy: -runnerEnergyCost, Happiness: 8},
	}
}
