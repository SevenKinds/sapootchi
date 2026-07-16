// Package assets embeds game art so builds are single, self-contained binaries
// (and WASM bundles). Decoding to renderer images happens in the game package.
//
// Layout: only files under assets/ ship in the binary. Reference material
// (brand kit, third-party packs) lives in /artpacks (gitignored); game-ready
// sprites are produced from it by `go run ./cmd/assetprep`.
package assets

import (
	"embed"
	_ "embed"
)

// BlobPNG is the classic pet sprite: the full/adult-size SAPO blob (green, big
// googly eyes). The baby is this same sprite rendered at 60% scale.
//
//go:embed sprites/blob.png
var BlobPNG []byte

// Sprites holds all prepared sprite art:
//
//	sprites/moods/sapo_NN.png — emotion poses (white bg removed from brand TIFs)
//	sprites/skins/<name>.png  — themed full-art skins (dress-up)
//
//go:embed sprites
var Sprites embed.FS
