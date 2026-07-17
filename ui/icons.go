package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Tab-bar icons are Nerd Font glyphs (Font Awesome 4.7 range — solidly present
// in MesloLGS NF). The old hand-drawn vector icons lived here; glyphs render
// crisper and match the care-button style.
type Icon rune

const (
	IconHome  Icon = '' // home
	IconGames Icon = '' // gamepad
	IconBag   Icon = '' // shopping-bag
	IconShop  Icon = '' // shopping-cart
	IconDress Icon = '' // paint-brush
	IconGear  Icon = '' // gear
)

// DrawIcon renders the icon centered at (cx, cy) in the given color.
func DrawIcon(dst *ebiten.Image, ic Icon, cx, cy float64, clr color.Color) {
	DrawGlyph(dst, rune(ic), cx, cy, 23, clr)
}
