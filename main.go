// Command sapootchi is the POC desktop/web build of the Loyalty Pet.
//
// The whole game is Go + Ebitengine so it can target WASM (web), iOS, Android,
// and desktop from one codebase. Game rules live in the Ebiten-free simulation
// package; this binary is just the renderer + input shell.
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/game"
)

func main() {
	// The framebuffer is rendered at 360x640 * ui.Scale (see game.Layout); the
	// window is a comfortable size Ebiten downscales that framebuffer into, and
	// is resizable so it can be enlarged without going blurry.
	const winW, winH = 450, 800
	ebiten.SetWindowSize(winW, winH)
	ebiten.SetWindowTitle("Sapootchi")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(game.New()); err != nil {
		log.Fatal(err)
	}
}
