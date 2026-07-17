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

// Sprites holds the prepared sprite art (moods/clouds/icons/games/items).
// Animations are NOT here: they embed on native builds only (anims_native.go)
// and are fetched over HTTP on web — they are the biggest asset block and
// would push the WASM bundle past Cloudflare Pages' 25 MiB file limit.
//
//go:embed sprites/blob.png sprites/moods sprites/clouds sprites/icons sprites/games sprites/items
var Sprites embed.FS

// IconFontTTF is MesloLGS Nerd Font Mono (Apache 2.0 / MIT patches) — used for
// its icon glyphs (Font Awesome & friends in the private-use area).
//
//go:embed fonts/MesloLGSNerdFontMono-Regular.ttf
var IconFontTTF []byte
