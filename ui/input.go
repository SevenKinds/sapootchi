package ui

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Pointer tracking in DESIGN-SPACE units (framebuffer coords / Scale). This lets
// the UI tell a tap from a swipe: a tap is a press+release with little movement;
// anything with more drag is a swipe and buttons must not fire. Touch is folded
// in so the same gestures work on mobile.
//
// Game.Update must call UpdateInput() once per frame before scenes run.

const tapSlop = 8.0 // max design-space drag distance still counted as a tap

var (
	held           bool
	pressX, pressY float64
	curX, curY     float64
	justPressed    bool
	justReleased   bool
	activeTouch    ebiten.TouchID
	touching       bool
)

// UpdateInput samples mouse + touch for this frame. Call once, first thing.
func UpdateInput() {
	justPressed, justReleased = false, false

	// Touch takes priority when present.
	if ids := inpututil.AppendJustPressedTouchIDs(nil); len(ids) > 0 && !touching {
		activeTouch = ids[0]
		touching = true
		x, y := ebiten.TouchPosition(activeTouch)
		curX, curY = float64(x)/Scale, float64(y)/Scale
		pressX, pressY = curX, curY
		held, justPressed = true, true
		return
	}
	if touching {
		if inpututil.IsTouchJustReleased(activeTouch) {
			touching, held, justReleased = false, false, true
			return
		}
		x, y := ebiten.TouchPosition(activeTouch)
		curX, curY = float64(x)/Scale, float64(y)/Scale
		return
	}

	mx, my := ebiten.CursorPosition()
	curX, curY = float64(mx)/Scale, float64(my)/Scale
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		held, justPressed = true, true
		pressX, pressY = curX, curY
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		held, justReleased = false, true
	}
}

// PointerHeld reports whether the pointer is currently down.
func PointerHeld() bool { return held }

// PointerJustPressed / PointerJustReleased report edges this frame.
func PointerJustPressed() bool  { return justPressed }
func PointerJustReleased() bool { return justReleased }

// Cursor returns the current pointer position in design-space units.
func Cursor() (float64, float64) { return curX, curY }

// PressPos returns where the current/last press began, in design-space units.
func PressPos() (float64, float64) { return pressX, pressY }

// DragDX is the horizontal drag from press to now (design-space units).
func DragDX() float64 { return curX - pressX }

func dragDist() float64 { return math.Hypot(curX-pressX, curY-pressY) }

func inRect(x, y, rx, ry, rw, rh float64) bool {
	return x >= rx && x <= rx+rw && y >= ry && y <= ry+rh
}

// Tapped reports a completed tap (press+release, little movement) inside the rect.
func Tapped(x, y, w, h float64) bool {
	if !justReleased || dragDist() > tapSlop {
		return false
	}
	return inRect(pressX, pressY, x, y, w, h) && inRect(curX, curY, x, y, w, h)
}
