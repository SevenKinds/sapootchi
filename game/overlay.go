package game

// The PayPal donate button is an HTML overlay the WASM canvas cannot draw
// itself. The game reserves a rectangle in the Settings layout and tells the
// browser to position the DOM button over it — but only while Settings is the
// visible page. On non-web builds this is all inert (see overlay_other.go /
// platform_other.go), so the same scene code runs everywhere.

// overlayRect is a design-space (360x640) rectangle handed to the JS bridge.
type overlayRect struct{ x, y, w, h float64 }

// settingsDonateRect reserves the donate slot (design-space), below the toggles.
func settingsDonateRect() overlayRect { return overlayRect{24, 300, ScreenW - 48, 52} }

// onSettingsPage reports whether Settings is the visible, settled top scene
// (no modal or mini-game pushed over it) — the only time overlays should show.
func (g *Game) onSettingsPage() bool {
	if len(g.scenes) != 1 {
		return false
	}
	m, ok := g.current().(*MainScene)
	return ok && m.settled() && m.idx == PageSettings
}

// syncOverlays shows/positions the donate button when Settings is up and hides
// it otherwise. Repositions every frame while shown so it tracks window resizes;
// only issues a hide on the transition away.
func (g *Game) syncOverlays() {
	on := g.onSettingsPage()
	if on {
		setWebOverlays(true, settingsDonateRect())
	} else if g.overlayShown {
		setWebOverlays(false, settingsDonateRect())
	}
	g.overlayShown = on
}
