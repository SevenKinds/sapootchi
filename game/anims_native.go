//go:build !js

package game

import (
	"io/fs"
	"path"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/assets"
)

// loadAnims (native): the animations are embedded; decode them synchronously.
func loadAnims(b *spriteBank) {
	dirs, err := fs.ReadDir(assets.Anims, "sprites/anims")
	if err != nil {
		return
	}
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		frameDir := path.Join("sprites/anims", d.Name())
		frames, err := fs.ReadDir(assets.Anims, frameDir)
		if err != nil {
			continue
		}
		var seq []*ebiten.Image
		for _, f := range frames { // ReadDir returns sorted names
			if !strings.HasSuffix(f.Name(), ".png") {
				continue
			}
			data, err := fs.ReadFile(assets.Anims, path.Join(frameDir, f.Name()))
			if err != nil {
				continue
			}
			seq = append(seq, decodeImage(data))
		}
		if len(seq) > 0 {
			b.Anims[d.Name()] = seq
		}
	}
}
