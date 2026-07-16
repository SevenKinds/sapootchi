package game

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/ui"
)

// Page is one swipeable screen hosted by MainScene. Pages draw into the area
// above the tab bar; navigation between them is the pager's job.
type Page interface {
	Icon() ui.Icon
	Label() string
	Update(g *Game) error
	Draw(g *Game, screen *ebiten.Image)
}

// Tab bar geometry (design-space units).
const (
	TabBarH  = 58.0
	PageH    = ScreenH - TabBarH // pages draw in 0..PageH
	swipeMin = 60.0              // min horizontal drag to change page
)

// MainScene hosts the pages as a horizontal swipeable pager with a single row
// of icon tab buttons at the bottom. Tap a tab or swipe the page to navigate.
// Mini-games are still pushed modally on top of this scene.
type MainScene struct {
	pages []Page
	idx   int     // settled page index
	pos   float64 // continuous position in page units (idx when settled)

	dragging  bool
	animating bool

	buf [2]*ebiten.Image // offscreen page buffers for the slide
}

func NewMainScene() *MainScene {
	return &MainScene{
		pages: []Page{
			&HomePage{},
			&GamesPage{},
			&InventoryPage{},
			&ShopPage{},
			&DressPage{},
			&SettingsPage{},
		},
	}
}

func (m *MainScene) settled() bool { return !m.dragging && !m.animating }

func (m *MainScene) Update(g *Game) error {
	m.updateSwipe()
	m.updateTabs()

	// Only the settled page takes interaction; Button taps already reject
	// swipes, but skipping Update during transit avoids edge cases entirely.
	if m.settled() {
		return m.pages[m.idx].Update(g)
	}
	return nil
}

func (m *MainScene) updateSwipe() {
	// Swipes must start on the page area, not the tab bar.
	if ui.PointerJustPressed() {
		_, py := ui.PressPos()
		m.dragging = py < PageH
	}
	if m.dragging && ui.PointerHeld() {
		m.pos = float64(m.idx) - ui.DragDX()/ScreenW
		m.pos = clampPos(m.pos, 0, float64(len(m.pages)-1))
		return
	}
	if m.dragging && ui.PointerJustReleased() {
		m.dragging = false
		if dx := ui.DragDX(); dx <= -swipeMin && m.idx < len(m.pages)-1 {
			m.idx++
		} else if dx >= swipeMin && m.idx > 0 {
			m.idx--
		}
		m.animating = true
	}
	if m.animating {
		target := float64(m.idx)
		m.pos += (target - m.pos) * 0.28
		if math.Abs(target-m.pos) < 0.002 {
			m.pos = target
			m.animating = false
		}
	}
}

func (m *MainScene) updateTabs() {
	for i := range m.pages {
		x, y, w, h := m.tabRect(i)
		if ui.Tapped(x, y, w, h) && i != m.idx {
			m.idx = i
			m.animating = true
		}
	}
}

func (m *MainScene) tabRect(i int) (x, y, w, h float64) {
	w = ScreenW / float64(len(m.pages))
	return float64(i) * w, PageH, w, TabBarH
}

func (m *MainScene) Draw(g *Game, screen *ebiten.Image) {
	m.drawPages(g, screen)
	m.drawTabBar(g, screen)
}

func (m *MainScene) drawPages(g *Game, screen *ebiten.Image) {
	if m.settled() {
		m.pages[m.idx].Draw(g, screen)
		return
	}

	// In transit: render the two neighboring pages into buffers and slide them.
	lo := int(math.Floor(m.pos))
	hi := int(math.Ceil(m.pos))
	if lo < 0 {
		lo = 0
	}
	if hi > len(m.pages)-1 {
		hi = len(m.pages) - 1
	}

	fw, fh := screen.Bounds().Dx(), screen.Bounds().Dy()
	for b := range m.buf {
		if m.buf[b] == nil || m.buf[b].Bounds().Dx() != fw || m.buf[b].Bounds().Dy() != fh {
			m.buf[b] = ebiten.NewImage(fw, fh)
		}
	}

	for n, i := range []int{lo, hi} {
		if n == 1 && hi == lo {
			break
		}
		buf := m.buf[n]
		buf.Clear()
		ui.BackgroundGradient(buf)
		m.pages[i].Draw(g, buf)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate((float64(i)-m.pos)*ScreenW*ui.Scale, 0)
		screen.DrawImage(buf, op)
	}
}

func (m *MainScene) drawTabBar(g *Game, screen *ebiten.Image) {
	// Bar background.
	ui.FillRoundRect(screen, 0, PageH, ScreenW, TabBarH, 0, ui.Track)
	ui.FillRoundRect(screen, 0, PageH, ScreenW, 2, 0, ui.PanelHi)

	for i, p := range m.pages {
		x, _, w, _ := m.tabRect(i)
		cx := x + w/2
		active := i == m.idx

		clr := ui.TextDim
		if active {
			clr = ui.Text
			// Active indicator pill behind the icon.
			ui.FillRoundRect(screen, float32(cx-19), float32(PageH+7), 38, 30, 15, ui.Panel)
		}
		ui.DrawIcon(screen, p.Icon(), cx, PageH+22, clr)
		if active {
			ui.DrawTextCenter(screen, p.Label(), cx, PageH+40, 9, ui.Text, true)
		}
	}
}

func clampPos(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
