//go:build js

package assets

import "embed"

// Anims is EMPTY on web builds: animations are fetched over HTTP at runtime
// to keep the WASM bundle under Cloudflare Pages' 25 MiB file limit.
var Anims embed.FS

// AnimNames lists the animation directories the web loader should fetch.
var AnimNames = []string{"coracoes", "estrelas", "sleepy", "triste", "wink"}
