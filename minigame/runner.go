package minigame

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"sapootchi/simulation"
	"sapootchi/ui"
)

// Runner is the endless-dodge mini-game and the DEDICATED energy-burner: it's
// exercise, so it spends a big chunk of energy (the way to calm an energized
// pet) and pays coins by distance. Jump the obstacles with space / up / click.
//
// Coordinates are in the 360x640 design space; the ui layer scales drawing.

const (
	runnerDurationTicks = 40 * 60 // hard cap ~40s
	runnerStartDelay    = 48      // "get ready" beat before it moves
	runnerGroundOffset  = 96
	runnerPetX          = 56
	runnerPetSize       = 42
	runnerObW           = 26
	runnerGravity       = 1.0
	runnerJumpV         = -15.0
	runnerBaseSpeed     = 4.2
	runnerSpeedRamp     = 0.0018
	runnerMaxSpeed      = 11.0

	runnerEnergyCost = 60.0 // the big burn
)

type obstacle struct {
	x      float64
	h      float64
	passed bool
}

// Runner implements minigame.Game.
type Runner struct {
	// Sprite, when set, is drawn as the player character instead of the shape
	// stand-in (the "real pet in mini-games" setting).
	Sprite *ebiten.Image

	w, h      float64
	petY      float64
	vy        float64
	onGround  bool
	obstacles []obstacle
	lastSpawn int
	scroll    float64
	score     int
	ticks     int
	done      bool
	crashed   bool
}

// NewRunner creates the game sized to the play area.
func NewRunner(width, height int) *Runner {
	r := &Runner{w: float64(width), h: float64(height)}
	r.petY = r.groundTop()
	r.onGround = true
	r.lastSpawn = -999
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

	// Jump (allowed during the ready beat so an eager tap still fires).
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

	// Spawn on a shrinking gap.
	gap := 96 - (r.ticks-runnerStartDelay)/110
	if gap < 46 {
		gap = 46
	}
	if r.ticks-r.lastSpawn >= gap {
		h := 32.0
		switch (r.ticks / gap) % 3 {
		case 0:
			h = 50.0 // taller block
		case 1:
			h = 40.0
		}
		r.obstacles = append(r.obstacles, obstacle{x: r.w + runnerObW, h: h})
		r.lastSpawn = r.ticks
	}

	alive := r.obstacles[:0]
	for _, o := range r.obstacles {
		o.x -= speed
		if !o.passed && o.x+runnerObW < runnerPetX {
			o.passed = true
			r.score++
		}
		if r.collides(o) {
			r.done = true
			r.crashed = true
		}
		if o.x+runnerObW > 0 {
			alive = append(alive, o)
		}
	}
	r.obstacles = alive

	if r.ticks >= runnerDurationTicks {
		r.done = true
	}
	return nil
}

func (r *Runner) jumpPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsKeyJustPressed(ebiten.KeyUp) ||
		inpututil.IsKeyJustPressed(ebiten.KeyW) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
		len(inpututil.AppendJustPressedTouchIDs(nil)) > 0
}

func (r *Runner) collides(o obstacle) bool {
	oTop := r.groundY() - o.h
	return runnerPetX < o.x+runnerObW &&
		runnerPetX+runnerPetSize > o.x &&
		r.petY+runnerPetSize > oTop
}

func (r *Runner) Draw(screen *ebiten.Image) {
	r.drawParallax(screen)

	gy := r.groundY()
	// Ground band + bright top edge.
	ui.FillRoundRect(screen, 0, float32(gy), float32(r.w), float32(runnerGroundOffset), 0, ui.Track)
	ui.FillRoundRect(screen, 0, float32(gy), float32(r.w), 4, 2, ui.PanelHi)
	// Scrolling ground dashes for a sense of speed.
	for x := -(int(r.scroll) % 34); x < int(r.w); x += 34 {
		ui.FillRoundRect(screen, float32(x), float32(gy+12), 16, 4, 2, ui.Panel)
	}

	// Obstacles.
	for _, o := range r.obstacles {
		ui.FillRoundRect(screen, float32(o.x), float32(gy-o.h), runnerObW, float32(o.h), 5, ui.Bad)
		ui.FillRoundRect(screen, float32(o.x), float32(gy-o.h), runnerObW, 5, 2.5,
			color.RGBA{0xff, 0x82, 0x78, 0xff})
	}

	r.drawPet(screen)

	// HUD.
	ui.DrawTextBold(screen, "RUNNER", 14, 14, 15, ui.Text)
	ui.DrawText(screen, "jump: space / up / tap", 14, 34, 11, ui.TextDim)
	scoreStr := "cleared " + ui.Itoa(r.score)
	ui.DrawTextBold(screen, scoreStr, r.w-14-ui.TextWidth(scoreStr, 15, true), 14, 15, ui.Gold)
	secs := (runnerDurationTicks - r.ticks) / 60
	tStr := ui.Itoa(secs) + "s"
	ui.DrawText(screen, tStr, r.w-14-ui.TextWidth(tStr, 12, false), 34, 12, ui.TextDim)

	if !r.running() {
		ui.DrawTextCenter(screen, "Get ready!", r.w/2, r.h/2-40, 26, ui.Text, true)
	}
}

// drawParallax draws faint background streaks that drift left with the run.
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

// drawPet draws the player character with velocity-based squash/stretch — the
// real blob sprite when Sprite is set, otherwise a green stand-in with eyes.
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

	// Two forward-looking googly eyes.
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

// Result: coins by distance, a big energy burn, a little happiness from play.
func (r *Runner) Result() Result {
	return Result{
		Score:     r.score,
		Coins:     r.score * 3,
		StatDelta: simulation.Stats{Energy: -runnerEnergyCost, Happiness: 8},
	}
}
