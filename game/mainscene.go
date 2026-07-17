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

// pointerCapturer is an optional Page extension: when a press lands where the
// page wants its own drag gesture (e.g. dragging an inventory item), the pager
// must not treat that gesture as a page swipe.
type pointerCapturer interface {
	CapturesPress(g *Game, x, y float64) bool
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

// Page indices (order in NewMainScene) — for programmatic navigation.
const (
	PageHome = iota
	PageGames
	PageItems
	PageShop
	PageDress
	PageSettings
)

func (m *MainScene) settled() bool { return !m.dragging && !m.animating }

// GoTo animates the pager to page i (e.g. Home's Feed shortcut -> Items).
func (m *MainScene) GoTo(i int) {
	if i >= 0 && i < len(m.pages) && i != m.idx {
		m.idx = i
		m.animating = true
	}
}

func (m *MainScene) Update(g *Game) error {
	m.updateSwipe(g)
	m.updateTabs()

	// Only the settled page takes interaction; Button taps already reject
	// swipes, but skipping Update during transit avoids edge cases entirely.
	if m.settled() {
		return m.pages[m.idx].Update(g)
	}
	return nil
}

func (m *MainScene) updateSwipe(g *Game) {
	// Swipes must start on the page area, not the tab bar — and not where the
	// current page wants its own drag gesture. NOTE: the capture check must run
	// BEFORE m.dragging is assigned (settled() reads it).
	if ui.PointerJustPressed() {
		px, py := ui.PressPos()
		wantSwipe := py < PageH
		if wantSwipe && m.settled() {
			if c, ok := m.pages[m.idx].(pointerCapturer); ok && c.CapturesPress(g, px, py) {
				wantSwipe = false
			}
		}
		m.dragging = wantSwipe
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
	// Bar background (themable nav palette).
	ui.FillRoundRect(screen, 0, PageH, ScreenW, TabBarH, 0, ui.NavBG)

	// Indicator pill slides with the pager position (tracks swipes live).
	tabW := ScreenW / float64(len(m.pages))
	pillCx := (m.pos + 0.5) * tabW
	ui.FillRoundRect(screen, float32(pillCx-19), float32(PageH+7), 38, 30, 15, ui.NavPill)

	for i, p := range m.pages {
		x, _, w, _ := m.tabRect(i)
		cx := x + w/2
		active := i == m.idx

		clr := ui.NavInkDim
		if active {
			clr = ui.NavInk
		}
		ui.DrawIcon(screen, p.Icon(), cx, PageH+22, clr)
		if active && m.settled() {
			ui.DrawTextCenter(screen, p.Label(), cx, PageH+40, 9, ui.NavInk, true)
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
