//go:build !js

package assets

import "embed"

// Anims embeds the reaction animations on native builds. On web they are
// fetched at runtime instead (see game's js anim loader).
//
//go:embed sprites/anims
var Anims embed.FS
