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

// Arrows is a rhythm/timing game (Arrowmania/DDR-style): arrows fall down four
// lanes toward a receptor row; hit the matching direction (arrow keys, or tap
// the lane) as the arrow crosses the receptors. Accuracy scores Perfect (3) or
// Good (1); good runs win Coffee.
//
// The chart is RANDOM per run (density ramps up, occasional double notes), so
// no two runs read the same. Inputs judge on PRESS, not release: a
// release-based tap would add latency a timing game can't afford.

const (
	arrowsDurationTicks = 32 * 60
	arrowsFallSpeed     = 3.4
	arrowsReceptorY     = 520.0
	arrowsPerfectWin    = 12.0 // px window around the receptor
	arrowsGoodWin       = 28.0

	arrowsHappiness  = 10.0
	arrowsEnergyCost = 12.0
)

// Coffee payout tiers by final score.
var arrowsCoffeeTiers = [3]int{50, 90, 130}

type fallingArrow struct {
	lane int
	y    float64
	hit  bool
}

// Arrows implements minigame.Game.
type Arrows struct {
	// Sprite, when set, draws the pet cheering beside the HUD.
	Sprite *ebiten.Image

	w, h      float64
	rng       *rand.Rand
	arrows    []fallingArrow
	nextSpawn int
	score     int
	combo     int
	best      int // best combo
	judge     string
	judgeC    color.RGBA
	judgeT    int
	ticks     int
	done      bool
}

// NewArrows creates the game sized to the play area.
func NewArrows(width, height int) *Arrows {
	return &Arrows{
		w:         float64(width),
		h:         float64(height),
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
		nextSpawn: 40,
	}
}

func (a *Arrows) Name() string { return "Arrows" }

func (a *Arrows) laneX(lane int) float64 {
	const laneW = 70.0
	x0 := (a.w - laneW*4) / 2
	return x0 + laneW/2 + float64(lane)*laneW
}

var arrowKeys = [4][]ebiten.Key{
	{ebiten.KeyLeft, ebiten.KeyA},
	{ebiten.KeyDown, ebiten.KeyS},
	{ebiten.KeyUp, ebiten.KeyW},
	{ebiten.KeyRight, ebiten.KeyD},
}

func (a *Arrows) Update() error {
	if a.done {
		return nil
	}
	a.ticks++
	if a.judgeT > 0 {
		a.judgeT--
	}

	// Random chart: beat gaps shrink as the run goes on; some beats rest, some
	// are DOUBLE notes (two lanes at once).
	a.nextSpawn--
	if a.nextSpawn <= 0 && a.ticks < arrowsDurationTicks-220 {
		if a.rng.Float64() > 0.12 { // ~1 in 8 beats rests
			lane := a.rng.Intn(4)
			a.arrows = append(a.arrows, fallingArrow{lane: lane, y: -30})
			if a.rng.Float64() < 0.12 && a.ticks > 600 { // doubles arrive later
				other := (lane + 1 + a.rng.Intn(3)) % 4
				a.arrows = append(a.arrows, fallingArrow{lane: other, y: -30})
			}
		}
		gap := 40 - a.ticks/140 // density ramp
		if gap < 22 {
			gap = 22
		}
		a.nextSpawn = gap + a.rng.Intn(10)
	}

	// Fall + miss detection.
	alive := a.arrows[:0]
	for _, ar := range a.arrows {
		ar.y += arrowsFallSpeed
		if !ar.hit && ar.y > arrowsReceptorY+arrowsGoodWin {
			a.setJudge("Miss", ui.Bad)
			a.combo = 0
			ar.hit = true // judged; keeps falling off-screen
		}
		if ar.y < a.h+40 {
			alive = append(alive, ar)
		}
	}
	a.arrows = alive

	// Inputs (press-time).
	for lane := 0; lane < 4; lane++ {
		if a.lanePressed(lane) {
			a.judgeLane(lane)
		}
	}

	if a.ticks >= arrowsDurationTicks {
		a.done = true
	}
	return nil
}

func (a *Arrows) lanePressed(lane int) bool {
	for _, k := range arrowKeys[lane] {
		if inpututil.IsKeyJustPressed(k) {
			return true
		}
	}
	// Touch/click: press in the lane column, lower half of the play area.
	if ui.PointerJustPressed() {
		px, py := ui.PressPos()
		if py > arrowsReceptorY-120 && px >= a.laneX(lane)-35 && px < a.laneX(lane)+35 {
			return true
		}
	}
	return false
}

// judgeLane scores the nearest unhit arrow in the lane, if inside a window.
func (a *Arrows) judgeLane(lane int) {
	bestIdx, bestDist := -1, 1e9
	for i, ar := range a.arrows {
		if ar.hit || ar.lane != lane {
			continue
		}
		d := ar.y - arrowsReceptorY
		if d < 0 {
			d = -d
		}
		if d < bestDist {
			bestDist, bestIdx = d, i
		}
	}
	if bestIdx < 0 || bestDist > arrowsGoodWin {
		return // empty lane press — no penalty, no reward
	}
	a.arrows[bestIdx].hit = true
	a.combo++
	if a.combo > a.best {
		a.best = a.combo
	}
	if bestDist <= arrowsPerfectWin {
		a.score += 3
		a.setJudge("PERFECT!", ui.Gold)
	} else {
		a.score++
		a.setJudge("Good", ui.Good)
	}
}

func (a *Arrows) setJudge(s string, c color.RGBA) {
	a.judge, a.judgeC, a.judgeT = s, c, 30
}

var arrowLaneColors = [4]color.RGBA{
	{0xe6, 0x56, 0x4a, 0xff}, // left: red
	{0x4c, 0x8d, 0xff, 0xff}, // down: blue
	{0x4c, 0xc9, 0x6d, 0xff}, // up: green
	{0xff, 0xb3, 0x3b, 0xff}, // right: yellow
}

func (a *Arrows) Draw(screen *ebiten.Image) {
	// Lane guides.
	for lane := 0; lane < 4; lane++ {
		x := a.laneX(lane)
		ui.StrokeLine(screen, float32(x-33), 60, float32(x-33), float32(a.h-40), 1, ui.PanelHi)
	}
	ui.StrokeLine(screen, float32(a.laneX(3)+33), 60, float32(a.laneX(3)+33), float32(a.h-40), 1, ui.PanelHi)

	// Receptors.
	for lane := 0; lane < 4; lane++ {
		x := a.laneX(lane)
		ui.FillCircle(screen, float32(x), arrowsReceptorY, 24, ui.PanelHi)
		ui.FillCircle(screen, float32(x), arrowsReceptorY, 20, ui.Track)
		drawChevron(screen, lane, x, arrowsReceptorY, ui.TextDim)
	}

	// Falling arrows.
	for _, ar := range a.arrows {
		if ar.hit {
			continue
		}
		x := a.laneX(ar.lane)
		c := arrowLaneColors[ar.lane]
		ui.FillCircle(screen, float32(x), float32(ar.y), 20, c)
		drawChevron(screen, ar.lane, x, ar.y, color.White)
	}

	// Judgment + combo.
	if a.judgeT > 0 {
		ui.DrawTextCenter(screen, a.judge, a.w/2, arrowsReceptorY-70, 17, a.judgeC, true)
	}
	if a.combo >= 3 {
		ui.DrawTextCenter(screen, "combo x"+ui.Itoa(a.combo), a.w/2, 64, 14, ui.Gold, true)
	}

	// The pet dances along at the bottom, bouncing on the beat.
	if a.Sprite != nil {
		bounce := math.Abs(math.Sin(float64(a.ticks)/8.5)) * -8
		ui.DrawImageFit(screen, a.Sprite, a.w/2-40, 562+bounce, 80, 68)
	}

	// HUD.
	ui.DrawTextBold(screen, "ARROWS", 14, 14, 15, ui.Text)
	ui.DrawText(screen, "arrow keys / tap lanes", 14, 34, 11, ui.TextDim)
	scoreStr := "score " + ui.Itoa(a.score)
	ui.DrawTextBold(screen, scoreStr, a.w-52-ui.TextWidth(scoreStr, 15, true), 14, 15, ui.Gold)
	secs := (arrowsDurationTicks - a.ticks) / 60
	tStr := ui.Itoa(secs) + "s"
	ui.DrawText(screen, tStr, a.w-52-ui.TextWidth(tStr, 12, false), 34, 12, ui.TextDim)
}

// drawChevron draws a direction chevron for the lane (L, D, U, R) at (x, y).
func drawChevron(dst *ebiten.Image, lane int, x, y float64, clr color.Color) {
	fx, fy := float32(x), float32(y)
	const s = 8
	switch lane {
	case 0: // left
		ui.StrokeLine(dst, fx+4, fy-s, fx-6, fy, 3, clr)
		ui.StrokeLine(dst, fx-6, fy, fx+4, fy+s, 3, clr)
	case 1: // down
		ui.StrokeLine(dst, fx-s, fy-4, fx, fy+6, 3, clr)
		ui.StrokeLine(dst, fx, fy+6, fx+s, fy-4, 3, clr)
	case 2: // up
		ui.StrokeLine(dst, fx-s, fy+4, fx, fy-6, 3, clr)
		ui.StrokeLine(dst, fx, fy-6, fx+s, fy+4, 3, clr)
	case 3: // right
		ui.StrokeLine(dst, fx-4, fy-s, fx+6, fy, 3, clr)
		ui.StrokeLine(dst, fx+6, fy, fx-4, fy+s, 3, clr)
	}
}

func (a *Arrows) Done() bool { return a.done }

// Result: coins by score, coffee at score tiers, happiness from dancing, and
// an energy cost (it IS exercise).
func (a *Arrows) Result() Result {
	coffee := 0
	for _, tier := range arrowsCoffeeTiers {
		if a.score >= tier {
			coffee++
		}
	}
	r := Result{
		Score:     a.score,
		Coins:     a.score / 3,
		StatDelta: simulation.Stats{Happiness: arrowsHappiness, Energy: -arrowsEnergyCost},
	}
	if coffee > 0 {
		r.Items = map[simulation.FoodKind]int{simulation.FoodCoffee: coffee}
	}
	return r
}
