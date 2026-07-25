//go:build js

package game

// OnWeb marks the WASM/browser build, where HTML overlays (ad + donate) exist.
// Settings draws its ad/donate slots only when this is true.
const OnWeb = true
