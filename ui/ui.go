// Package ui is the shared look-and-feel layer: palette, anti-aliased fonts,
// and drawing primitives (rounded rects, buttons, stat bars, gradient bg).
//
// Both the game and the mini-games import it, so visuals stay consistent and
// there is a single place to retheme. It imports ebiten but no game logic.
package ui

import (
	"bytes"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
)

// Palette — a friendly dark theme.
var (
	BGTop    = color.RGBA{0x3b, 0x46, 0x63, 0xff} // top gradient (blue)
	BGBottom = color.RGBA{0x20, 0x25, 0x33, 0xff} // bottom gradient (deep)
	Panel    = color.RGBA{0x2a, 0x30, 0x40, 0xff}
	PanelHi  = color.RGBA{0x33, 0x3a, 0x4d, 0xff}

	Text    = color.RGBA{0xed, 0xef, 0xf5, 0xff}
	TextDim = color.RGBA{0x9a, 0xa3, 0xb2, 0xff}

	Accent      = color.RGBA{0x4c, 0x8d, 0xff, 0xff} // primary buttons (blue)
	AccentHover = color.RGBA{0x6b, 0xa5, 0xff, 0xff}
	Secondary   = color.RGBA{0x7c, 0x5c, 0xff, 0xff} // nav buttons (purple)
	SecondaryHi = color.RGBA{0x95, 0x7a, 0xff, 0xff}
	Disabled    = color.RGBA{0x44, 0x4a, 0x59, 0xff}

	Good  = color.RGBA{0x4c, 0xc9, 0x6d, 0xff} // stat > 50
	Warn  = color.RGBA{0xff, 0xb3, 0x3b, 0xff} // 25-50
	Bad   = color.RGBA{0xe6, 0x56, 0x4a, 0xff} // < 25
	Energy = color.RGBA{0x46, 0xc7, 0xe0, 0xff} // energy is special (inverted)
	Gold   = color.RGBA{0xff, 0xd2, 0x4c, 0xff} // coins
	Track  = color.RGBA{0x1c, 0x20, 0x2b, 0xff} // stat-bar background
	Shadow = color.RGBA{0x00, 0x00, 0x00, 0x40} // soft drop shadow
)

// Scale is the render-density multiplier. Scenes author coordinates and font
// sizes in a fixed 360x640 design space; every ui draw call multiplies by Scale
// so the actual framebuffer is rendered at higher resolution (crisp, not an
// upscale of a small image). Input coordinates are divided back out.
var Scale = 2.0

var (
	regular *text.GoTextFaceSource
	bold    *text.GoTextFaceSource
	white   *ebiten.Image // 1x1 source for triangle fills
)

func init() {
	r, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		log.Fatalf("load regular font: %v", err)
	}
	b, err := text.NewGoTextFaceSource(bytes.NewReader(gobold.TTF))
	if err != nil {
		log.Fatalf("load bold font: %v", err)
	}
	regular, bold = r, b

	white = ebiten.NewImage(1, 1)
	white.Fill(color.White)
}

func face(size float64, isBold bool) *text.GoTextFace {
	src := regular
	if isBold {
		src = bold
	}
	return &text.GoTextFace{Source: src, Size: size}
}

// DrawText draws left-aligned text with its top-left at (x, y).
func DrawText(dst *ebiten.Image, s string, x, y, size float64, clr color.Color) {
	drawText(dst, s, x, y, size, clr, false)
}

// DrawTextBold draws bold left-aligned text.
func DrawTextBold(dst *ebiten.Image, s string, x, y, size float64, clr color.Color) {
	drawText(dst, s, x, y, size, clr, true)
}

// DrawTextCenter draws text horizontally centered on cx.
func DrawTextCenter(dst *ebiten.Image, s string, cx, y, size float64, clr color.Color, isBold bool) {
	f := face(size, isBold)
	w, _ := text.Measure(s, f, size*1.2)
	drawText(dst, s, cx-w/2, y, size, clr, isBold)
}

func drawText(dst *ebiten.Image, s string, x, y, size float64, clr color.Color, isBold bool) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x*Scale, y*Scale)
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(dst, s, face(size*Scale, isBold), op)
}

// FillRoundRect fills a rounded rectangle, composed from primitives (no path
// plumbing needed).
func FillRoundRect(dst *ebiten.Image, x, y, w, h, r float32, clr color.Color) {
	s := float32(Scale)
	x, y, w, h, r = x*s, y*s, w*s, h*s, r*s
	if r > w/2 {
		r = w / 2
	}
	if r > h/2 {
		r = h / 2
	}
	vector.DrawFilledRect(dst, x+r, y, w-2*r, h, clr, true)
	vector.DrawFilledRect(dst, x, y+r, w, h-2*r, clr, true)
	vector.DrawFilledCircle(dst, x+r, y+r, r, clr, true)
	vector.DrawFilledCircle(dst, x+w-r, y+r, r, clr, true)
	vector.DrawFilledCircle(dst, x+r, y+h-r, r, clr, true)
	vector.DrawFilledCircle(dst, x+w-r, y+h-r, r, clr, true)
}

// DrawImageFit draws img scaled to fit inside the design-space box (x,y,w,h),
// centered, preserving aspect ratio.
func DrawImageFit(dst *ebiten.Image, img *ebiten.Image, x, y, w, h float64) {
	iw := float64(img.Bounds().Dx())
	ih := float64(img.Bounds().Dy())
	s := w / iw
	if hs := h / ih; hs < s {
		s = hs
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(s*Scale, s*Scale)
	op.GeoM.Translate((x+(w-iw*s)/2)*Scale, (y+(h-ih*s)/2)*Scale)
	op.Filter = ebiten.FilterLinear
	dst.DrawImage(img, op)
}

// DrawImageStretch draws img stretched to exactly fill the design-space box —
// used for squash/stretch animation where distortion is the point.
func DrawImageStretch(dst *ebiten.Image, img *ebiten.Image, x, y, w, h float64) {
	iw := float64(img.Bounds().Dx())
	ih := float64(img.Bounds().Dy())
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(w/iw*Scale, h/ih*Scale)
	op.GeoM.Translate(x*Scale, y*Scale)
	op.Filter = ebiten.FilterLinear
	dst.DrawImage(img, op)
}

// FillCircle fills a circle centered at (cx, cy) in design-space units.
func FillCircle(dst *ebiten.Image, cx, cy, r float32, clr color.Color) {
	s := float32(Scale)
	vector.DrawFilledCircle(dst, cx*s, cy*s, r*s, clr, true)
}

// StrokeLine draws a line in design-space units.
func StrokeLine(dst *ebiten.Image, x0, y0, x1, y1, width float32, clr color.Color) {
	s := float32(Scale)
	vector.StrokeLine(dst, x0*s, y0*s, x1*s, y1*s, width*s, clr, true)
}

// BackgroundGradient paints the whole screen with the vertical theme gradient.
func BackgroundGradient(dst *ebiten.Image) {
	h := dst.Bounds().Dy()
	w := float32(dst.Bounds().Dx())
	const bands = 48
	bandH := float32(h) / bands
	for i := 0; i < bands; i++ {
		t := float32(i) / (bands - 1)
		c := lerp(BGTop, BGBottom, t)
		vector.DrawFilledRect(dst, 0, float32(i)*bandH, w, bandH+1, c, false)
	}
}

func lerp(a, b color.RGBA, t float32) color.RGBA {
	return color.RGBA{
		uint8(float32(a.R) + (float32(b.R)-float32(a.R))*t),
		uint8(float32(a.G) + (float32(b.G)-float32(a.G))*t),
		uint8(float32(a.B) + (float32(b.B)-float32(a.B))*t),
		0xff,
	}
}
