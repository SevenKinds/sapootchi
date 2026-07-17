package game

import "github.com/hajimehoshi/ebiten/v2"

// animDelivery carries an asynchronously loaded animation into the game loop
// (web builds fetch animations over HTTP; the map is only written here).
type animDelivery struct {
	name   string
	frames []*ebiten.Image
}

// drainAnims applies any finished background loads. Called from Game.Update so
// the Anims map is never written concurrently with reads.
func (g *Game) drainAnims() {
	if g.Sprites.animsIn == nil {
		return
	}
	for {
		select {
		case d := <-g.Sprites.animsIn:
			g.Sprites.Anims[d.name] = d.frames
		default:
			return
		}
	}
}
