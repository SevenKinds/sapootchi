package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Button is a rounded, hover-aware clickable rect with a centered label.
type Button struct {
	X, Y, W, H float64
	Label      string
	Secondary  bool // purple (navigation) instead of blue (primary)
}

// Hover reports whether the cursor is over the button.
func (b Button) Hover() bool {
	cx, cy := Cursor()
	return inRect(cx, cy, b.X, b.Y, b.W, b.H)
}

// Clicked reports a completed tap on the button this frame. A press that drags
// (a swipe) does not fire — pages rely on this to coexist with swiping.
func (b Button) Clicked() bool {
	return Tapped(b.X, b.Y, b.W, b.H)
}

// Draw renders the button. Disabled buttons are greyed and flat.
func (b Button) Draw(dst *ebiten.Image, enabled bool) {
	fill := Accent
	if b.Secondary {
		fill = Secondary
	}
	if enabled && b.Hover() {
		fill = AccentHover
		if b.Secondary {
			fill = SecondaryHi
		}
	}
	if !enabled {
		fill = Disabled
	}

	const r = 10
	// Soft drop shadow.
	FillRoundRect(dst, float32(b.X), float32(b.Y+2), float32(b.W), float32(b.H),
		r, Shadow)
	FillRoundRect(dst, float32(b.X), float32(b.Y), float32(b.W), float32(b.H), r, fill)

	label := ButtonInk
	if !enabled {
		label = TextDim
	}
	DrawTextCenter(dst, b.Label, b.X+b.W/2, b.Y+b.H/2-8, 15, label, true)
}

// GlyphButton is a quieter button: panel-colored, an icon glyph on top and a
// small label under it. Used for the Home care row — present, not shouty.
type GlyphButton struct {
	X, Y, W, H float64
	Glyph      rune
	GlyphColor color.Color
	Label      string
}

// Hover reports whether the cursor is over the button.
func (b GlyphButton) Hover() bool {
	cx, cy := Cursor()
	return inRect(cx, cy, b.X, b.Y, b.W, b.H)
}

// Clicked reports a completed tap on the button this frame.
func (b GlyphButton) Clicked() bool {
	return Tapped(b.X, b.Y, b.W, b.H)
}

// Draw renders the button. Disabled buttons grey out entirely.
func (b GlyphButton) Draw(dst *ebiten.Image, enabled bool) {
	bg := Panel
	if enabled && b.Hover() {
		bg = PanelHi
	}
	FillRoundRect(dst, float32(b.X), float32(b.Y), float32(b.W), float32(b.H), 12, bg)

	gc, lc := b.GlyphColor, TextDim
	if !enabled {
		gc, lc = Disabled, Disabled
	}
	DrawGlyph(dst, b.Glyph, b.X+b.W/2, b.Y+b.H/2-8, 19, gc)
	DrawTextCenter(dst, b.Label, b.X+b.W/2, b.Y+b.H-17, 9.5, lc, true)
}

// StatBar draws a labeled meter (0..100) with a rounded track and fill.
// A custom fill color may be given (e.g. Energy); pass nil to auto-color by value.
func StatBar(dst *ebiten.Image, label string, value, x, y, w float64, fixed color.Color) {
	const h = 14
	FillRoundRect(dst, float32(x), float32(y), float32(w), h, h/2, Track)

	fillW := w * value / 100
	if fillW < h { // keep the rounded cap visible even near-empty
		fillW = 0
		if value > 0 {
			fillW = h
		}
	}
	clr := fixed
	if clr == nil {
		clr = statColor(value)
	}
	if fillW > 0 {
		FillRoundRect(dst, float32(x), float32(y), float32(fillW), h, h/2, clr)
	}

	DrawTextBold(dst, label, x, y-16, 12, TextDim)
	valStr := itoa(int(value+0.5)) + "%"
	tw := measure(valStr, face(12, true))
	DrawText(dst, valStr, x+w-tw, y-16, 12, TextDim)
}

func statColor(v float64) color.RGBA {
	switch {
	case v < 25:
		return Bad
	case v < 50:
		return Warn
	default:
		return Good
	}
}
