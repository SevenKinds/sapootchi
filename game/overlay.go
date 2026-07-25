package game

// HTML overlays (ad iframe, PayPal donate button) that the WASM canvas cannot
// draw itself. The game reserves rectangles in the Settings layout and tells the
// browser to position matching DOM elements over them — but only while Settings
// is the visible page. On non-web builds this is all inert (see overlay_other.go
// / platform_other.go), so the same scene code runs everywhere.

// overlayRect is a design-space (360x640) rectangle handed to the JS bridge.
type overlayRect struct{ x, y, w, h float64 }

// Reserved slots in the Settings screen (design-space). Kept below the toggles
// and the dev row so nothing overlaps.
func settingsDonateRect() overlayRect { return overlayRect{24, 400, ScreenW - 48, 48} }
func settingsAdRect() overlayRect     { return overlayRect{24, 462, ScreenW - 48, 90} }

// onSettingsPage reports whether Settings is the visible, settled top scene
// (no modal or mini-game pushed over it) — the only time overlays should show.
func (g *Game) onSettingsPage() bool {
	if len(g.scenes) != 1 {
		return false
	}
	m, ok := g.current().(*MainScene)
	return ok && m.settled() && m.idx == PageSettings
}

// syncOverlays shows/positions the web overlays when Settings is up and hides
// them otherwise. Repositions every frame while shown so it tracks window
// resizes; only issues a hide on the transition away.
func (g *Game) syncOverlays() {
	on := g.onSettingsPage()
	if on {
		setWebOverlays(true, settingsAdRect(), settingsDonateRect())
	} else if g.overlayShown {
		setWebOverlays(false, settingsAdRect(), settingsDonateRect())
	}
	g.overlayShown = on
}
