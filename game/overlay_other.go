//go:build !js

package game

// setWebOverlays is a no-op off the web (no DOM to overlay). See overlay_js.go.
func setWebOverlays(show bool, donate overlayRect) {}
