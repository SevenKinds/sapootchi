// Package assets embeds game art so builds are single, self-contained binaries
// (and WASM bundles). Decoding to renderer images happens in the game package.
package assets

import _ "embed"

// Layout: only files under assets/ referenced by an //go:embed directive ship
// in the binary. Reference material (brand kit, third-party packs) lives in
// /artpacks, which is gitignored — it is source material, not runtime data.
//
// BlobPNG is the pet sprite: the full/adult-size blob (green, big googly eyes).
// The baby is this same sprite rendered at 60% scale — there is no baby asset.
//
//go:embed sprites/blob.png
var BlobPNG []byte
