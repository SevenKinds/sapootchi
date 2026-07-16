package game

import "github.com/hajimehoshi/ebiten/v2"

// Scene is one screen. The Game holds a stack of them; the top scene receives
// Update/Draw. Scenes: Home, MiniGame, Shop, DressUp. Ebiten imposes no
// structure — this stack is ours.
type Scene interface {
	Update(g *Game) error
	Draw(g *Game, screen *ebiten.Image)
}
