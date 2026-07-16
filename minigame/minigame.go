// Package minigame defines the plug-in interface every mini-game implements, so
// each game is a self-contained plug-in. A game may pay out Items OR Coins (or
// both) plus a stat delta — do not assume one currency per game.
//
// POC ships catch-food only. Adding a game later = implement Game and register it.
package minigame

import (
	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/simulation"
)

// Result is what a finished mini-game pays out. The caller applies it to the pet.
type Result struct {
	Score     int
	Coins     int
	StatDelta simulation.Stats             // additive delta to visible stats
	Items     map[simulation.FoodKind]int // e.g. catch-food awards 1-3 food items
}

// Game is one playable mini-game. The MiniGame scene drives it: Update each
// tick, Draw each frame, and once Done, read Result and apply it to the pet.
type Game interface {
	Name() string
	Update() error
	Draw(screen *ebiten.Image)
	Done() bool
	Result() Result
}
