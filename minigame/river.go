package minigame

import (
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"sapootchi/simulation"
	"sapootchi/ui"
)

// River is the vertical 3-lane dodge-and-collect game: the frog swims upstream
// while rocks float down — switch lanes to avoid them and grab the coins.
// Three lives; speed ramps. Coins collected pay out DIRECTLY.
//
// Controls: tap the left/right half of the screen (or arrow keys / A-D).
// Uses real randomness: lane patterns vary per run.

const (
	riverDurationTicks = 40 * 60
	riverLanes         = 3
	riverLaneW         = 96.0
	riverPetY          = 520.0
	riverPetSize       = 46.0
	riverRockSize      = 40.0
	riverCoinR         = 12.0
	riverLives         = 3
	riverInvulnTicks   = 70

	riverEnergyCost = 18.0
)

type riverThing struct {
	lane int
	y    float64
	coin bool
	dead bool
}

// River implements minigame.Game.
type River struct {
	// Sprite, when set, is drawn as the swimming pet.
	Sprite *ebiten.Image

	w, h    float64
	rng     *rand.Rand
	lane    int
	things  []riverThing
	nextRow int
	lives   int
	invuln  int
	coins   int
	dodged  int
	scroll  float64
	ticks   int
	done    bool
}

// NewRiver creates the game sized to the play area.
func NewRiver(width, height int) *River {
	return &River{
		w:       float64(width),
		h:       float64(height),
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
		lane:    1,
		lives:   riverLives,
		nextRow: 60,
	}
}

func (r *River) Name() string { return "River" }

func (r *River) laneX(lane int) float64 {
	x0 := (r.w - riverLaneW*riverLanes) / 2
	return x0 + riverLaneW/2 + float64(lane)*riverLaneW
}

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

	r.handleInput()

	// Spawn rows: 1-2 rocks, never blocking every lane; coins in free lanes.
	r.nextRow--
	if r.nextRow <= 0 {
		r.spawnRow()
		gap := 56 - r.ticks/160
		if gap < 30 {
			gap = 30
		}
		r.nextRow = gap + r.rng.Intn(14)
	}

	// Move things; judge collisions at the pet row.
	speed := r.speed()
	alive := r.things[:0]
	for _, t := range r.things {
		t.y += speed
		hitBand := t.y > riverPetY-riverPetSize/2 && t.y < riverPetY+riverPetSize/2
		if !t.dead && hitBand && t.lane == r.lane {
			t.dead = true
			if t.coin {
				r.coins++
			} else if r.invuln == 0 {
				r.lives--
				r.invuln = riverInvulnTicks
				if r.lives <= 0 {
					r.done = true
				}
			}
		}
		if !t.dead && !t.coin && t.y > riverPetY+riverPetSize {
			r.dodged++
			t.dead = true
		}
		if !t.dead && t.y < r.h+riverRockSize {
			alive = append(alive, t)
		}
	}
	r.things = alive

	if r.ticks >= riverDurationTicks {
		r.done = true
	}
	return nil
}

func (r *River) handleInput() {
	left := inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA)
	right := inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyD)
	if ui.PointerJustPressed() {
		px, py := ui.PressPos()
		if py > 60 { // below the HUD/quit row
			if px < r.w/2 {
				left = true
			} else {
				right = true
			}
		}
	}
	if left && r.lane > 0 {
		r.lane--
	}
	if right && r.lane < riverLanes-1 {
		r.lane++
	}
}

func (r *River) spawnRow() {
	rocks := 1
	if r.ticks > 900 && r.rng.Float64() < 0.4 {
		rocks = 2 // later on, two-rock rows (one lane always free)
	}
	lanes := r.rng.Perm(riverLanes)
	for i := 0; i < rocks; i++ {
		r.things = append(r.things, riverThing{lane: lanes[i], y: -riverRockSize})
	}
	// Coin in one of the free lanes, most of the time.
	if r.rng.Float64() < 0.7 {
		free := lanes[rocks:]
		r.things = append(r.things, riverThing{
			lane: free[r.rng.Intn(len(free))],
			y:    -riverRockSize - 60,
			coin: true,
		})
	}
}

func (r *River) Draw(screen *ebiten.Image) {
	// River banks + lanes.
	x0 := (r.w - riverLaneW*riverLanes) / 2
	ui.FillRoundRect(screen, float32(x0-8), 0, float32(riverLaneW*riverLanes+16), float32(r.h), 0,
		color.RGBA{0x1f, 0x33, 0x45, 0xff}) // water
	for i := 0; i <= riverLanes; i++ {
		x := x0 + float64(i)*riverLaneW
		ui.StrokeLine(screen, float32(x), 0, float32(x), float32(r.h), 1, color.RGBA{0xff, 0xff, 0xff, 0x16})
	}
	// Flow streaks moving DOWN (the world scrolls past the swimming frog).
	for i := 0; i < riverLanes; i++ {
		x := r.laneX(i)
		off := int(r.scroll) % 120
		for y := -120 + off; y < int(r.h); y += 120 {
			ui.FillRoundRect(screen, float32(x-2+float64(i*7%11)), float32(y), 4, 26, 2,
				color.RGBA{0xff, 0xff, 0xff, 0x12})
		}
	}

	// Things.
	for _, t := range r.things {
		if t.dead {
			continue
		}
		x := r.laneX(t.lane)
		if t.coin {
			ui.FillCircle(screen, float32(x), float32(t.y), riverCoinR, ui.Gold)
			ui.FillCircle(screen, float32(x), float32(t.y), riverCoinR/2, color.RGBA{0xc9, 0x9e, 0x1f, 0xff})
		} else {
			ui.FillRoundRect(screen, float32(x-riverRockSize/2), float32(t.y-riverRockSize/2),
				riverRockSize, riverRockSize, 13, color.RGBA{0x8a, 0x93, 0xa5, 0xff})
			ui.FillRoundRect(screen, float32(x-riverRockSize/2+6), float32(t.y-riverRockSize/2+6),
				riverRockSize-22, riverRockSize-26, 8, color.RGBA{0x6d, 0x76, 0x88, 0xff})
		}
	}

	// Pet (blinks while invulnerable).
	if r.invuln == 0 || (r.ticks/6)%2 == 0 {
		x := r.laneX(r.lane)
		wob := math.Sin(float64(r.ticks)/10) * 3
		if r.Sprite != nil {
			ui.DrawImageFit(screen, r.Sprite, x-riverPetSize/2+wob, riverPetY-riverPetSize/2, riverPetSize, riverPetSize)
		} else {
			ui.FillRoundRect(screen, float32(x-riverPetSize/2+wob), float32(riverPetY-riverPetSize/2),
				riverPetSize, riverPetSize, 16, ui.Good)
		}
	}

	// HUD.
	ui.DrawTextBold(screen, "RIVER", 14, 14, 15, ui.Text)
	ui.DrawText(screen, "tap left/right — dodge rocks, grab coins", 14, 34, 11, ui.TextDim)
	coinStr := "coins " + ui.Itoa(r.coins)
	ui.DrawTextBold(screen, coinStr, r.w-52-ui.TextWidth(coinStr, 15, true), 14, 15, ui.Gold)
	// Lives as dots.
	for i := 0; i < riverLives; i++ {
		c := ui.PanelHi
		if i < r.lives {
			c = ui.Bad
		}
		ui.FillCircle(screen, float32(r.w-60-float64(i)*18), 44, 5, c)
	}
	secs := (riverDurationTicks - r.ticks) / 60
	tStr := ui.Itoa(secs) + "s"
	ui.DrawText(screen, tStr, 14, 52, 12, ui.TextDim)
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
