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

// River: the frog swims upstream with FREE horizontal movement (follow the
// pointer, or arrows/A-D) — no lanes. Dodge what the river brings: rocks
// (animated), wide logs, and driftwood that slides sideways. Grab the coins.
// Three lives; speed ramps. Coins pay out directly.
//
// Uses real randomness: patterns vary per run.

const (
	riverDurationTicks = 40 * 60
	riverBankMargin    = 26.0 // unswimmable banks left/right
	riverPetY          = 520.0
	riverPetSize       = 46.0
	riverCoinR         = 12.0
	riverLives         = 3
	riverInvulnTicks   = 70
	riverFollowSpeed   = 6.5 // max px/tick toward the pointer

	riverEnergyCost = 18.0
)

type riverKind int

const (
	rockThing riverKind = iota
	logThing
	driftThing
	coinThing
)

type riverThing struct {
	kind    riverKind
	x, y    float64
	w       float64 // collision width (logs are wide)
	vx      float64 // driftwood slides
	variant int     // rock art variant
	dead    bool
	counted bool // dodged bookkeeping (things stay VISIBLE until off-screen)
}

// River implements minigame.Game.
type River struct {
	// Injected art.
	Sprite     *ebiten.Image     // the pet, when the real-sprite setting is on
	Rocks      [][]*ebiten.Image // animated rock variants (frame lists)
	Foam       []*ebiten.Image   // animated foam patches (water texture)
	CoinFrames []*ebiten.Image   // the game's spinning coin (pickups)

	w, h      float64
	rng       *rand.Rand
	petX      float64
	mouseIdle bool // keys silence the mouse until it is clicked again
	things    []riverThing
	nextRow   int
	lives     int
	invuln    int
	coins     int
	dodged    int
	scroll    float64
	ticks     int
	done      bool
}

// NewRiver creates the game sized to the play area.
func NewRiver(width, height int) *River {
	return &River{
		w:       float64(width),
		h:       float64(height),
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
		petX:    float64(width) / 2,
		lives:   riverLives,
		nextRow: 60,
	}
}

func (r *River) Name() string { return "River" }

func (r *River) speed() float64 {
	s := 3.2 + float64(r.ticks)*0.0016
	if s > 7.2 {
		s = 7.2
	}
	return s
}

func (r *River) Update() error {
	if r.done {
		return nil
	}
	r.ticks++
	if r.invuln > 0 {
		r.invuln--
	}
	r.scroll += r.speed()

	r.move()

	// Spawning.
	r.nextRow--
	if r.nextRow <= 0 {
		r.spawn()
		gap := 52 - r.ticks/170
		if gap < 26 {
			gap = 26
		}
		r.nextRow = gap + r.rng.Intn(14)
	}

	// Drift things down (and sideways) and judge collisions.
	speed := r.speed()
	alive := r.things[:0]
	for _, t := range r.things {
		t.y += speed
		if t.kind == driftThing {
			t.x += t.vx
			if t.x < riverBankMargin+t.w/2 || t.x > r.w-riverBankMargin-t.w/2 {
				t.vx = -t.vx
			}
		}
		if !t.dead && r.hits(t) {
			t.dead = true
			if t.kind == coinThing {
				r.coins++
			} else if r.invuln == 0 {
				r.lives--
				r.invuln = riverInvulnTicks
				if r.lives <= 0 {
					r.done = true
				}
			}
		}
		if !t.dead && !t.counted && t.kind != coinThing && t.y > riverPetY+riverPetSize {
			r.dodged++
			t.counted = true // keep drifting — despawn only past the screen edge
		}
		if !t.dead && t.y < r.h+80 {
			alive = append(alive, t)
		}
	}
	r.things = alive

	if r.ticks >= riverDurationTicks {
		r.done = true
	}
	return nil
}

// move: free horizontal control — glide toward the pointer, or key steering.
// Playing with the keys puts the mouse to sleep (no drift toward a stale
// cursor); clicking wakes it again.
func (r *River) move() {
	left := ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA)
	right := ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD)
	if left || right {
		r.mouseIdle = true
	}
	if ui.PointerJustPressed() {
		r.mouseIdle = false
	}

	target := r.petX
	switch {
	case left:
		target = r.petX - riverFollowSpeed
	case right:
		target = r.petX + riverFollowSpeed
	case ui.PointerHeld():
		target, _ = ui.Cursor()
	case !r.mouseIdle:
		if mx, _ := ui.Cursor(); mx > 0 {
			target = mx // desktop: follow the mouse like catch-food
		}
	}
	d := clampF(target-r.petX, -riverFollowSpeed, riverFollowSpeed)
	r.petX = clampF(r.petX+d, riverBankMargin+riverPetSize/2, r.w-riverBankMargin-riverPetSize/2)
}

// spawn drops the next thing in: rock, log, driftwood, or coin — always
// leaving swimmable water somewhere.
func (r *River) spawn() {
	usable := r.w - 2*riverBankMargin
	roll := r.rng.Float64()
	switch {
	case roll < 0.14 && r.ticks > 500:
		// Log: wide, spans up to 60% of the river — squeeze past its end.
		w := usable * (0.42 + r.rng.Float64()*0.18)
		left := r.rng.Float64() < 0.5
		x := riverBankMargin + w/2
		if !left {
			x = r.w - riverBankMargin - w/2
		}
		r.things = append(r.things, riverThing{kind: logThing, x: x, y: -40, w: w})
	case roll < 0.30 && r.ticks > 900:
		// Driftwood: slides sideways while coming down.
		vx := 0.8 + r.rng.Float64()*0.9
		if r.rng.Float64() < 0.5 {
			vx = -vx
		}
		r.things = append(r.things, riverThing{
			kind: driftThing, w: 52, vx: vx, y: -40,
			x: riverBankMargin + 30 + r.rng.Float64()*(usable-60),
		})
	default:
		// Rock: 1-2 of them at free positions with a guaranteed gap.
		n := 1
		if r.ticks > 800 && r.rng.Float64() < 0.45 {
			n = 2
		}
		var first float64
		for i := 0; i < n; i++ {
			x := riverBankMargin + 24 + r.rng.Float64()*(usable-48)
			if i == 1 && math.Abs(x-first) < 110 {
				x = math.Mod(first+usable/2, usable-48) + riverBankMargin + 24
			}
			first = x
			r.things = append(r.things, riverThing{
				kind: rockThing, x: x, y: -40, w: 42,
				variant: r.rng.Intn(maxInt(1, len(r.Rocks))),
			})
		}
	}
	// Coin, most rows.
	if r.rng.Float64() < 0.6 {
		r.things = append(r.things, riverThing{
			kind: coinThing, w: riverCoinR * 2, y: -110,
			x: riverBankMargin + 20 + r.rng.Float64()*(usable-40),
		})
	}
}

// hits tests the pet (circle-ish) against a thing (width box around its y).
func (r *River) hits(t riverThing) bool {
	if t.y < riverPetY-riverPetSize/2-14 || t.y > riverPetY+riverPetSize/2+14 {
		return false
	}
	return math.Abs(t.x-r.petX) < (t.w+riverPetSize)/2*0.82
}

func (r *River) Draw(screen *ebiten.Image) {
	// Water + banks.
	ui.FillRoundRect(screen, float32(riverBankMargin-8), 0, float32(r.w-2*riverBankMargin+16), float32(r.h), 0,
		color.RGBA{0x1f, 0x33, 0x45, 0xff})
	ui.StrokeLine(screen, float32(riverBankMargin-8), 0, float32(riverBankMargin-8), float32(r.h), 2, ui.PanelHi)
	ui.StrokeLine(screen, float32(r.w-riverBankMargin+8), 0, float32(r.w-riverBankMargin+8), float32(r.h), 2, ui.PanelHi)

	// Water texture: animated foam patches — a foam seam along each bank plus
	// sparse drifting patches mid-water for depth.
	if len(r.Foam) > 0 {
		f := r.Foam[(r.ticks/9)%len(r.Foam)]
		fw := float64(f.Bounds().Dx())
		// Bank seams (half off the water edge), scrolling with the flow.
		const seamScale = 0.55
		step := fw * seamScale * 0.8
		off := math.Mod(r.scroll*0.9, step)
		for y := -step + off; y < r.h+step; y += step {
			ui.DrawImageNearest(screen, f, riverBankMargin-fw*seamScale*0.72, y, seamScale, 0.30)
			ui.DrawImageNearest(screen, f, r.w-riverBankMargin-fw*seamScale*0.28, y, seamScale, 0.30)
		}
		// Drifting mid-water patches.
		for i := 0; i < 3; i++ {
			scale := 0.5 + float64(i)*0.14
			span := r.h + fw*scale
			y := math.Mod(r.scroll*(0.65+float64(i)*0.12)+float64(i)*260, span) - fw*scale
			x := riverBankMargin + 30 + float64(i*97%int(r.w-2*riverBankMargin-90))
			ui.DrawImageNearest(screen, f, x, y, scale, 0.10)
		}
	}

	// Things.
	for _, t := range r.things {
		if t.dead {
			continue
		}
		switch t.kind {
		case coinThing:
			drawSpinCoin(screen, r.CoinFrames, r.ticks, t.x, t.y, riverCoinR*2)
		case rockThing:
			if len(r.Rocks) > 0 {
				frames := r.Rocks[t.variant%len(r.Rocks)]
				f := frames[(r.ticks/8)%len(frames)]
				fw := float64(f.Bounds().Dx())
				s := (t.w * 1.9) / fw // rock art has water ripple margins
				ui.DrawImageNearest(screen, f, t.x-fw*s/2, t.y-fw*s/2, s, 1)
			} else {
				ui.FillCircle(screen, float32(t.x), float32(t.y), float32(t.w/2), color.RGBA{0x8a, 0x93, 0xa5, 0xff})
			}
		case logThing:
			log := color.RGBA{0x8b, 0x5e, 0x3c, 0xff}
			ui.FillRoundRect(screen, float32(t.x-t.w/2), float32(t.y-13), float32(t.w), 26, 13, log)
			ui.FillCircle(screen, float32(t.x-t.w/2+13), float32(t.y), 9, color.RGBA{0xb0, 0x82, 0x5a, 0xff})
			ui.FillCircle(screen, float32(t.x-t.w/2+13), float32(t.y), 4, color.RGBA{0x6d, 0x49, 0x2e, 0xff})
		case driftThing:
			drift := color.RGBA{0xa1, 0x6f, 0x47, 0xff}
			ui.FillRoundRect(screen, float32(t.x-t.w/2), float32(t.y-9), float32(t.w), 18, 9, drift)
			ui.StrokeLine(screen, float32(t.x-t.w/2+8), float32(t.y), float32(t.x+t.w/2-8), float32(t.y), 2,
				color.RGBA{0x6d, 0x49, 0x2e, 0xff})
		}
	}

	// Pet (blinks while invulnerable), with a little swim wobble.
	if r.invuln == 0 || (r.ticks/6)%2 == 0 {
		wob := math.Sin(float64(r.ticks)/10) * 3
		if r.Sprite != nil {
			ui.DrawImageFit(screen, r.Sprite, r.petX-riverPetSize/2+wob, riverPetY-riverPetSize/2, riverPetSize, riverPetSize)
		} else {
			ui.FillRoundRect(screen, float32(r.petX-riverPetSize/2+wob), float32(riverPetY-riverPetSize/2),
				riverPetSize, riverPetSize, 16, ui.Good)
		}
	}

	// HUD.
	ui.DrawTextBold(screen, "RIVER", 14, 14, 15, ui.Text)
	ui.DrawText(screen, "swim freely — dodge everything", 14, 34, 11, ui.TextDim)
	coinStr := "coins " + ui.Itoa(r.coins)
	ui.DrawTextBold(screen, coinStr, r.w-52-ui.TextWidth(coinStr, 15, true), 14, 15, ui.Gold)
	for i := 0; i < riverLives; i++ {
		c := ui.PanelHi
		if i < r.lives {
			c = ui.Bad
		}
		ui.FillCircle(screen, float32(r.w-60-float64(i)*18), 44, 5, c)
	}
	secs := (riverDurationTicks - r.ticks) / 60
	ui.DrawText(screen, ui.Itoa(secs)+"s", 14, 52, 12, ui.TextDim)
}

func (r *River) Done() bool { return r.done }

// Result: collected coins pay out directly (plus a little for distance),
// proper exercise energy cost, small happiness.
func (r *River) Result() Result {
	return Result{
		Score:     r.coins,
		Coins:     r.coins*2 + r.dodged/4,
		StatDelta: simulation.Stats{Energy: -riverEnergyCost, Happiness: 6},
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
