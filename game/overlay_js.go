//go:build js

package game

import "syscall/js"

// setWebOverlays drives the JS bridge in web/index.html (window.sapootchiOverlay)
// to place or hide the ad + donate DOM elements over the canvas.
func setWebOverlays(show bool, ad, donate overlayRect) {
	o := js.Global().Get("sapootchiOverlay")
	if !o.Truthy() {
		return
	}
	if !show {
		o.Call("hide")
		return
	}
	o.Call("place", "ad-slot", ad.x, ad.y, ad.w, ad.h)
	o.Call("place", "donate-slot", donate.x, donate.y, donate.w, donate.h)
}
