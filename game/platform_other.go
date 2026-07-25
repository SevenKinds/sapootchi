//go:build !js

package game

// OnWeb is false off the web: no HTML overlays, so Settings skips the ad/donate
// slots. See platform_js.go.
const OnWeb = false
