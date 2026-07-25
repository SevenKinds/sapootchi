package game

import (
	"image"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/ui"
)

// DressPage is the skins tab: a grid of the brand's themed SAPO looks. Tap a
// tile to equip; "Classic" is the default blob. The equipped skin persists.
// The grid scrolls (drag or wheel) since there are many looks.
type DressPage struct {
	flash      string
	flashUntil int

	scroll float64 // vertical grid offset (>=0 scrolls content up)
	grab   bool    // a scroll drag is in progress
	lastY  float64 // cursor y last frame, for drag deltas
}

func (p *DressPage) Icon() ui.Icon { return ui.IconDress }
func (p *DressPage) Label() string { return "Dress" }

// Grid layout (design-space).
const (
	dressGridY = 236.0
	dressCols  = 4
	dressTile  = 74.0
	dressGap   = 8.0
	dressVpTop = 224.0 // grid viewport starts here (below the live preview)
)

// skinAt maps tile index -> skin name ("" = classic, index 0).
func (p *DressPage) skinAt(g *Game, i int) (name string, ok bool) {
	if i == 0 {
		return "", true
	}
	if i-1 < len(g.Sprites.SkinNames) {
		return g.Sprites.SkinNames[i-1], true
	}
	return "", false
}

func (p *DressPage) tileRect(i int) (x, y, w, h float64) {
	total := dressCols*dressTile + (dressCols-1)*dressGap
	x0 := (ScreenW - total) / 2
	col, row := i%dressCols, i/dressCols
	return x0 + float64(col)*(dressTile+dressGap), dressGridY + float64(row)*(dressTile+dressGap), dressTile, dressTile
}

// maxScroll is how far the grid can scroll so the last row rests at the bottom.
func (p *DressPage) maxScroll(g *Game) float64 {
	n := 1 + len(g.Sprites.SkinNames)
	rows := (n + dressCols - 1) / dressCols
	contentBottom := dressGridY + float64(rows)*(dressTile+dressGap)
	m := contentBottom - PageH + 8
	if m < 0 {
		m = 0
	}
	return m
}

// CapturesPress claims presses in the grid viewport so the pager scrolls the
// grid instead of paging (swipe from the header/preview to change tabs).
func (p *DressPage) CapturesPress(g *Game, x, y float64) bool {
	return y >= dressVpTop && y < PageH
}

func (p *DressPage) Update(g *Game) error {
	maxScroll := p.maxScroll(g)

	// Scroll: mouse wheel (desktop) + drag (touch) within the viewport.
	if _, wy := ebiten.Wheel(); wy != 0 {
		p.scroll -= wy * 24
	}
	_, cy := ui.Cursor()
	if ui.PointerJustPressed() {
		_, py := ui.PressPos()
		p.grab = py >= dressVpTop && py < PageH
		p.lastY = cy
	}
	if p.grab && ui.PointerHeld() {
		p.scroll -= cy - p.lastY
		p.lastY = cy
	}
	if ui.PointerJustReleased() {
		p.grab = false
	}
	p.scroll = clampPos(p.scroll, 0, maxScroll)

	// Equip taps (scroll-adjusted; ignore tiles outside the viewport).
	n := 1 + len(g.Sprites.SkinNames)
	for i := 0; i < n; i++ {
		x, y, w, h := p.tileRect(i)
		y -= p.scroll
		if y+h < dressVpTop || y > PageH {
			continue
		}
		if !ui.Tapped(x, y, w, h) {
			continue
		}
		name, ok := p.skinAt(g, i)
		if !ok || g.Pet.Skin == name {
			continue
		}
		if !g.OwnsSkin(name) {
			p.flash = displayName(name) + " is locked — unlock it in the Shop (" + ui.Itoa(skinPrice) + "c)"
			p.flashUntil = g.tick + 150
			continue
		}
		g.Pet.Skin = name // skins are PER PET — dressing the active one
		g.Save()
		label := "Classic"
		if name != "" {
			label = displayName(name)
		}
		p.flash = g.Pet.Name + " equipped " + label + "!"
		p.flashUntil = g.tick + 150
	}
	return nil
}

func (p *DressPage) Draw(g *Game, screen *ebiten.Image) {
	ui.DrawTextBold(screen, "Dress Up", 24, 28, 24, ui.Text)
	ui.DrawText(screen, "dressing "+g.Pet.Name+" — switch pets on Home", 24, 62, 12, ui.TextDim)

	// Live preview: the active pet exactly as it looks right now.
	g.DrawBlob(screen, ScreenW/2, 148)

	ui.DrawText(screen, "tap a look to equip it", 24, 208, 12, ui.TextDim)

	// The grid draws into a clipped viewport so scrolled tiles don't spill over
	// the header/preview.
	vp := screen.SubImage(image.Rect(0, int(dressVpTop*ui.Scale), screen.Bounds().Dx(), int(PageH*ui.Scale))).(*ebiten.Image)
	n := 1 + len(g.Sprites.SkinNames)
	for i := 0; i < n; i++ {
		name, _ := p.skinAt(g, i)
		x, y, w, h := p.tileRect(i)
		y -= p.scroll
		if y+h < dressVpTop || y > PageH {
			continue
		}

		if g.Pet.Skin == name {
			ui.FillRoundRect(vp, float32(x-3), float32(y-3), float32(w+6), float32(h+6), 14, ui.Accent)
		}
		owned := g.OwnsSkin(name)
		ui.FillRoundRect(vp, float32(x), float32(y), float32(w), float32(h), 12,
			colIf(owned, ui.PanelHi, ui.Panel))

		img := g.Sprites.Blob
		if name != "" {
			img = g.Sprites.Skins[name]
		}
		if owned {
			ui.DrawImageFit(vp, img, x+6, y+6, w-12, h-12)
		} else {
			// Locked: dimmed art + a lock glyph.
			ui.DrawImageFitAlpha(vp, img, x+6, y+6, w-12, h-12, 0.22)
			ui.DrawGlyph(vp, '\uf023', x+w/2, y+h/2, 20, ui.TextDim)
		}
	}

	// Scrollbar hint on the right edge when there's overflow.
	if maxScroll := p.maxScroll(g); maxScroll > 0 {
		trackH := PageH - dressVpTop - 8
		thumbH := trackH * trackH / (trackH + maxScroll)
		thumbY := dressVpTop + 4 + (trackH-thumbH)*p.scroll/maxScroll
		ui.FillRoundRect(screen, ScreenW-8, float32(dressVpTop+4), 3, float32(trackH), 1.5, ui.Track)
		ui.FillRoundRect(screen, ScreenW-8, float32(thumbY), 3, float32(thumbH), 1.5, ui.TextDim)
	}

	if g.tick < p.flashUntil {
		ui.DrawTextCenter(screen, p.flash, ScreenW/2, 214, 12, ui.Gold, true)
	}
}

// displayName prettifies a skin name via the pose-name table.
func displayName(name string) string {
	if name == "" {
		return "Classic"
	}
	if d, ok := skinDisplay[name]; ok {
		return d
	}
	return strings.ToUpper(name[:1]) + name[1:]
}
