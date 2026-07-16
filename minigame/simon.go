package minigame

import (
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/simulation"
	"sapootchi/ui"
)

// Simon is "playing together" -> feeds Happiness (and hidden Intelligence).
// Watch the pads light up, then repeat the sequence. Three strikes ends the
// game; each completed round grows the sequence by one.
//
// Unlike the other games, Simon NEEDS real randomness — a deterministic
// sequence could be memorized across plays — so it owns a seeded rng.

const (
	simonMaxRounds = 20
	simonStrikes   = 3

	simonShowTicks = 26 // how long a pad stays lit during playback
	simonGapTicks  = 12 // gap between playback lights
	simonHappiness = 15.0
	simonEnergyCst = 8.0
)

type simonState int

const (
	simonPlayback simonState = iota
	simonInput
	simonDone
)

// Simon implements minigame.Game.
type Simon struct {
	// Sprite, when set, draws the pet watching the game.
	Sprite *ebiten.Image

	w, h float64
	rng  *rand.Rand

	seq     []int
	state   simonState
	playIdx int // which sequence step is being shown
	phase   int // tick counter within playback step
	inIdx   int // how much of the sequence the player has repeated
	strikes int
	flash   int // >0: pad index+1 flashing from player tap
	flashT  int
	wrongT  int // "wrong!" feedback timer
	score   int
	ticks   int
}

// NewSimon creates the game sized to the play area.
func NewSimon(width, height int) *Simon {
	s := &Simon{
		w:   float64(width),
		h:   float64(height),
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	s.seq = []int{s.rng.Intn(4)}
	return s
}

func (s *Simon) Name() string { return "Simon" }

// padRect returns the rect of pad i (2x2 grid).
func (s *Simon) padRect(i int) (x, y, w, h float64) {
	const size, gap = 120.0, 14.0
	gridW := size*2 + gap
	x0 := (s.w - gridW) / 2
	y0 := 240.0
	col, row := float64(i%2), float64(i/2)
	return x0 + col*(size+gap), y0 + row*(size+gap), size, size
}

var simonPadColors = [4]color.RGBA{
	{0x4c, 0xc9, 0x6d, 0xff}, // green
	{0xe6, 0x56, 0x4a, 0xff}, // red
	{0xff, 0xb3, 0x3b, 0xff}, // yellow
	{0x4c, 0x8d, 0xff, 0xff}, // blue
}

func (s *Simon) Update() error {
	if s.state == simonDone {
		return nil
	}
	s.ticks++
	if s.flashT > 0 {
		s.flashT--
	}
	if s.wrongT > 0 {
		s.wrongT--
	}

	switch s.state {
	case simonPlayback:
		s.phase++
		if s.phase >= simonShowTicks+simonGapTicks {
			s.phase = 0
			s.playIdx++
			if s.playIdx >= len(s.seq) {
				s.state = simonInput
				s.inIdx = 0
			}
		}
	case simonInput:
		for i := 0; i < 4; i++ {
			if !ui.Tapped(s.padRect(i)) {
				continue
			}
			s.flash, s.flashT = i+1, 14
			if i == s.seq[s.inIdx] {
				s.inIdx++
				if s.inIdx == len(s.seq) {
					s.score++
					if s.score >= simonMaxRounds {
						s.state = simonDone
						return nil
					}
					s.seq = append(s.seq, s.rng.Intn(4))
					s.playIdx, s.phase = 0, -30 // brief pause before playback
					s.state = simonPlayback
				}
			} else {
				s.strikes++
				s.wrongT = 40
				if s.strikes >= simonStrikes {
					s.state = simonDone
					return nil
				}
				// Replay the sequence from the top.
				s.playIdx, s.phase = 0, -30
				s.state = simonPlayback
			}
			break
		}
	}
	return nil
}

func (s *Simon) litPad() int {
	if s.state == simonPlayback && s.playIdx < len(s.seq) && s.phase >= 0 && s.phase < simonShowTicks {
		return s.seq[s.playIdx]
	}
	if s.flashT > 0 && s.flash > 0 {
		return s.flash - 1
	}
	return -1
}

func (s *Simon) Draw(screen *ebiten.Image) {
	// The pet watches from above the pads.
	if s.Sprite != nil {
		ui.DrawImageFit(screen, s.Sprite, s.w/2-55, 96, 110, 96)
	}

	lit := s.litPad()
	for i := 0; i < 4; i++ {
		x, y, w, h := s.padRect(i)
		c := simonPadColors[i]
		if i != lit { // dim unlit pads
			c = color.RGBA{c.R / 3, c.G / 3, c.B / 3, 0xff}
		}
		ui.FillRoundRect(screen, float32(x), float32(y), float32(w), float32(h), 18, c)
	}

	// HUD.
	ui.DrawTextBold(screen, "SIMON", 14, 14, 15, ui.Text)
	ui.DrawText(screen, "repeat the sequence", 14, 34, 11, ui.TextDim)
	scoreStr := "round " + ui.Itoa(s.score+1)
	ui.DrawTextBold(screen, scoreStr, s.w-52-ui.TextWidth(scoreStr, 15, true), 14, 15, ui.Gold)

	// Strikes as dots.
	for i := 0; i < simonStrikes; i++ {
		c := ui.PanelHi
		if i < s.strikes {
			c = ui.Bad
		}
		ui.FillCircle(screen, float32(s.w-60-float64(i)*18), 44, 5, c)
	}

	switch {
	case s.wrongT > 0:
		ui.DrawTextCenter(screen, "Wrong — watch again!", s.w/2, 208, 13, ui.Bad, true)
	case s.state == simonPlayback:
		ui.DrawTextCenter(screen, "Watch...", s.w/2, 208, 13, ui.TextDim, true)
	case s.state == simonInput:
		ui.DrawTextCenter(screen, "Your turn!", s.w/2, 208, 13, ui.Text, true)
	}
}

func (s *Simon) Done() bool { return s.state == simonDone }

// Result: happiness from playing together, hidden intelligence + coins by how
// far the sequence got, and a tiny energy cost.
func (s *Simon) Result() Result {
	return Result{
		Score:     s.score,
		Coins:     s.score * 2,
		StatDelta: simulation.Stats{Happiness: simonHappiness, Energy: -simonEnergyCst},
		Hidden:    simulation.Hidden{Intelligence: float64(s.score) * 2},
	}
}
