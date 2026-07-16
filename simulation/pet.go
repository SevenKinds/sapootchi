// Package simulation holds all pet game logic as plain data.
//
// RULE: this package must have ZERO Ebiten (or any renderer) imports. Pet state
// is plain data; decay is a pure function of elapsed wall-clock time; rendering
// is a pure view of state built elsewhere. This boundary keeps the logic
// unit-testable, server-portable (for hard currency later), and
// renderer-agnostic. Do not leak ebiten.* in here.
package simulation

import (
	"math/rand"
	"time"
)

// Phase is the pet's life stage. The egg phase is near-instant: a new pet
// hatches immediately, so the first real phase is Baby. Branching evolutions
// (Cute/Smart/Wild -> Royal/Wizard/Ninja) are stubbed for the POC.
type Phase int

const (
	PhaseEgg Phase = iota // transient; hatches immediately into Baby
	PhaseBaby
	PhaseCute
	PhaseSmart
	PhaseWild
)

func (p Phase) String() string {
	switch p {
	case PhaseEgg:
		return "Egg"
	case PhaseBaby:
		return "Baby"
	case PhaseCute:
		return "Cute"
	case PhaseSmart:
		return "Smart"
	case PhaseWild:
		return "Wild"
	default:
		return "Unknown"
	}
}

// RenderScale is the sprite scale for the phase. Baby = the full blob sprite at
// 60%; grown forms render at full size. (Kept here as plain data; the renderer
// reads it, but it imports nothing.)
func (p Phase) RenderScale() float64 {
	if p == PhaseBaby {
		return 0.60
	}
	return 1.0
}

// Personality is assigned at hatch (random) for the POC. Same mechanics per
// personality — it only flavors animation + dialogue. Emergent personality is v2.
type Personality int

const (
	PersonalityLazy Personality = iota
	PersonalityHyper
	PersonalityGreedy
	PersonalityCurious
)

func (p Personality) String() string {
	switch p {
	case PersonalityLazy:
		return "Lazy"
	case PersonalityHyper:
		return "Hyper"
	case PersonalityGreedy:
		return "Greedy"
	case PersonalityCurious:
		return "Curious"
	default:
		return "Unknown"
	}
}

// Stats are the four visible meters, each 0..100.
type Stats struct {
	Happiness float64
	Hunger    float64
	Hygiene   float64
	Energy    float64
}

// Hidden stats drive evolution and flavor; never shown as bars.
type Hidden struct {
	Intelligence float64
	Friendship   float64
}

// Pet is the entire game state. It is plain, serializable data.
type Pet struct {
	Name        string
	Phase       Phase
	Personality Personality
	Stats       Stats
	Hidden      Hidden
	Coins       int
	Inventory   map[FoodKind]int

	// Skin is the equipped cosmetic look ("" = classic). Plain data — the
	// renderer resolves the name to art.
	Skin string

	BornAt   time.Time
	LastSeen time.Time // reference point for wall-clock decay / away rolls

	// Soft-fail state: hunger hitting 0 makes the pet run away. Leaving food
	// out gives a per-day chance he returns.
	Away        bool
	FoodLeftOut bool

	// Asleep: forced sleep triggered when Energy hits 0 (see energy.go). Stays
	// asleep until Energy regenerates to energyWakeThreshold.
	Asleep bool
}

// NewPet creates a freshly hatched baby. The egg is skipped (hatches instantly).
//
// Lore: you found him just in the nick of time — starving (hunger nearly at the
// runaway point), grubby, and lonely. A fresh start is too healthy; the first
// session is a rescue: feed him, clean him up, cheer him up. The two starter
// apples are exactly enough to pull him out of danger.
func NewPet(name string, personality Personality, now time.Time) *Pet {
	return &Pet{
		Name:        name,
		Phase:       PhaseBaby,
		Personality: personality,
		Stats:       Stats{Happiness: 25, Hunger: 12, Hygiene: 20, Energy: 55},
		Coins:       0,
		Inventory:   map[FoodKind]int{FoodApple: 2},
		BornAt:      now,
		LastSeen:    now,
	}
}

// NewRandomPet hatches a baby with a random personality.
func NewRandomPet(name string, now time.Time, rng *rand.Rand) *Pet {
	return NewPet(name, Personality(rng.Intn(4)), now)
}

func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}
