package minigame

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/simulation"
	"sapootchi/ui"
)

// Scrub is bath time -> feeds Hygiene. Dirt spots appear on the pet; rub the
// pointer back and forth over them to scrub them off. Cleaned spots pop into
// bubbles. Gentle exercise: a small energy cost.
//
// Coordinates are design-space; deterministic spot placement comes from tick
// hashing (no rand needed).

const (
	scrubDurationTicks = 25 * 60
	scrubSpotMax       = 4   // max dirt spots alive at once
	scrubSpawnEvery    = 100 // ticks between spawn attempts
	scrubSpotR         = 22.0
	scrubDirtFull      = 100.0
	scrubPowerFactor   = 1.4 // dirt removed per design-px of pointer travel

	scrubEnergyCost  = 10.0
	scrubHygieneBase = 40.0
)

type dirtSpot struct {
	x, y float64
	dirt float64 // 0 = clean
}

type bubble struct {
	x, y, vy float64
	life     int
}

// Scrub implements minigame.Game.
type Scrub struct {
	// Sprite, when set, draws the real pet in the tub.
	Sprite *ebiten.Image

	w, h    float64
	spots   []dirtSpot
	bubbles []bubble
	lastX   float64
	lastY   float64
	hadPos  bool
	score   int
	ticks   int
	done    bool
}

// NewScrub creates the game sized to the play area.
func NewScrub(width, height int) *Scrub {
	return &Scrub{w: float64(width), h: float64(height)}
}

func (s *Scrub) Name() string { return "Scrub" }

// petBox is the area the pet occupies (and where dirt can appear).
func (s *Scrub) petBox() (x, y, w, h float64) {
	w, h = 220, 190
	return (s.w - w) / 2, 200, w, h
}

func (s *Scrub) Update() error {
	if s.done {
		return nil
	}
	s.ticks++

	// Spawn dirt on a cadence, deterministically placed via tick hashing.
	if s.ticks%scrubSpawnEvery == 1 && len(s.aliveSpots()) < scrubSpotMax {
		px, py, pw, ph := s.petBox()
		hx := float64((s.ticks*131)%1000) / 1000
		hy := float64((s.ticks*577)%1000) / 1000
		// Keep spots inside an ellipse so they sit "on" the blob.
		ex := (hx*2 - 1) * 0.72
		ey := (hy*2 - 1) * 0.62
		if ex*ex+ey*ey > 1 {
			n := math.Hypot(ex, ey)
			ex, ey = ex/n*0.7, ey/n*0.7
		}
		s.spots = append(s.spots, dirtSpot{
			x:    px + pw/2 + ex*pw/2,
			y:    py + ph/2 + ey*ph/2,
			dirt: scrubDirtFull,
		})
	}

	s.updateScrubbing()
	s.updateBubbles()

	if s.ticks >= scrubDurationTicks {
		s.done = true
	}
	return nil
}

func (s *Scrub) updateScrubbing() {
	if !ui.PointerHeld() {
		s.hadPos = false
		return
	}
	cx, cy := ui.Cursor()
	if !s.hadPos {
		s.lastX, s.lastY = cx, cy
		s.hadPos = true
		return
	}
	travel := math.Hypot(cx-s.lastX, cy-s.lastY)
	s.lastX, s.lastY = cx, cy
	if travel <= 0 {
		return
	}

	for i := range s.spots {
		sp := &s.spots[i]
		if sp.dirt <= 0 {
			continue
		}
		if math.Hypot(cx-sp.x, cy-sp.y) > scrubSpotR {
			continue
		}
		sp.dirt -= travel * scrubPowerFactor
		// Foam while scrubbing.
		if s.ticks%4 == 0 {
			s.bubbles = append(s.bubbles, bubble{x: cx, y: cy, vy: -0.6, life: 30})
		}
		if sp.dirt <= 0 {
			s.score++
			s.popBubbles(sp.x, sp.y)
		}
	}
}

func (s *Scrub) popBubbles(x, y float64) {
	for i := 0; i < 7; i++ {
		a := float64(i) / 7 * 2 * math.Pi
		s.bubbles = append(s.bubbles, bubble{
			x: x + math.Cos(a)*8, y: y + math.Sin(a)*8,
			vy: -1.2 - float64(i%3)*0.4, life: 45,
		})
	}
}

func (s *Scrub) updateBubbles() {
	alive := s.bubbles[:0]
	for _, b := range s.bubbles {
		b.y += b.vy
		b.life--
		if b.life > 0 {
			alive = append(alive, b)
		}
	}
	s.bubbles = alive
}

func (s *Scrub) aliveSpots() []dirtSpot {
	out := s.spots[:0:0]
	for _, sp := range s.spots {
		if sp.dirt > 0 {
			out = append(out, sp)
		}
	}
	return out
}

func (s *Scrub) Draw(screen *ebiten.Image) {
	px, py, pw, ph := s.petBox()

	// Tub: a rounded basin under the pet.
	ui.FillRoundRect(screen, float32(px-30), float32(py+ph-40), float32(pw+60), 70, 24, ui.PanelHi)

	// Pet.
	if s.Sprite != nil {
		ui.DrawImageFit(screen, s.Sprite, px, py, pw, ph)
	} else {
		ui.FillRoundRect(screen, float32(px+20), float32(py+30), float32(pw-40), float32(ph-40), 60, ui.Good)
	}

	// Dirt spots (shrink as they get scrubbed).
	for _, sp := range s.spots {
		if sp.dirt <= 0 {
			continue
		}
		r := float32(scrubSpotR * (0.4 + 0.6*sp.dirt/scrubDirtFull))
		ui.FillCircle(screen, float32(sp.x), float32(sp.y), r, color.RGBA{0x7a, 0x58, 0x38, 0xb4})
		ui.FillCircle(screen, float32(sp.x-4), float32(sp.y-4), r*0.45, color.RGBA{0x63, 0x47, 0x2c, 0xb4})
	}

	// Bubbles.
	for _, b := range s.bubbles {
		alpha := uint8(160 * b.life / 45)
		ui.FillCircle(screen, float32(b.x), float32(b.y), 5, color.RGBA{0xff, 0xff, 0xff, alpha})
	}

	// HUD.
	ui.DrawTextBold(screen, "SCRUB", 14, 14, 15, ui.Text)
	ui.DrawText(screen, "rub the dirt off!", 14, 34, 11, ui.TextDim)
	scoreStr := "cleaned " + ui.Itoa(s.score)
	ui.DrawTextBold(screen, scoreStr, s.w-52-ui.TextWidth(scoreStr, 15, true), 14, 15, ui.Gold)
	secs := (scrubDurationTicks - s.ticks) / 60
	tStr := ui.Itoa(secs) + "s"
	ui.DrawText(screen, tStr, s.w-52-ui.TextWidth(tStr, 12, false), 34, 12, ui.TextDim)
}

func (s *Scrub) Done() bool { return s.done }

// Result: hygiene from the bath plus a bonus per spot, small coins, gentle
// energy cost.
func (s *Scrub) Result() Result {
	return Result{
		Score:     s.score,
		Coins:     s.score * 2,
		StatDelta: simulation.Stats{Hygiene: scrubHygieneBase + float64(s.score)*5, Energy: -scrubEnergyCost},
	}
}
