package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Hand-authored 16x16 pixel sprites for items with no pack art (coffee, pill),
// drawn to match the brackeys-fruit style: chunky pixels, dark outline.

func buildPixelSprite(rows []string, pal map[rune]color.RGBA) *ebiten.Image {
	img := ebiten.NewImage(16, 16)
	for y, row := range rows {
		for x, ch := range row {
			if c, ok := pal[ch]; ok {
				img.Set(x, y, c)
			}
		}
	}
	return img
}

var pixelPal = map[rune]color.RGBA{
	'D': {0x2b, 0x1d, 0x12, 0xff}, // dark outline
	'W': {0xf4, 0xf0, 0xe8, 0xff}, // cream / white
	'w': {0xd9, 0xd2, 0xc4, 0xff}, // cream shade
	'B': {0x6f, 0x4a, 0x2f, 0xff}, // coffee
	'b': {0x8a, 0x5f, 0x3e, 0xff}, // coffee light
	'S': {0xb9, 0xc4, 0xc9, 0xaa}, // steam
	'C': {0x46, 0xc7, 0xe0, 0xff}, // capsule cyan
	'c': {0x2e, 0x9c, 0xb4, 0xff}, // capsule cyan shade
	'H': {0xff, 0xff, 0xff, 0xff}, // highlight
	'G': {0xff, 0xd2, 0x4c, 0xff}, // gold spark
}

var coffeeArt = []string{
	"................",
	"....S.....S.....",
	"...S.....S......",
	"....S.....S.....",
	"................",
	"..DDDDDDDDDD....",
	".DWWWWWWWWWWD...",
	".DWBBBBBBBBWDDD.",
	".DWBbBBBBBBWD.D.",
	".DWBBBBBBBBWD.D.",
	".DWWWWWWWWWWDDD.",
	".DwWWWWWWWWwD...",
	"..DwWWWWWWwD....",
	"...DDDDDDDD.....",
	"..DwwwwwwwwD....",
	"...DDDDDDDD.....",
}

var pillArt = []string{
	"................",
	"..........GG....",
	"..........GG....",
	".....DDD........",
	"....DWWWD.......",
	"...DWWWHWD......",
	"..DWWWWWWCD.....",
	".DWWWWWCCCcD....",
	".DWWWCCCCCcD....",
	".DWCCCCCCCcD....",
	".DcCCCCCCcD.....",
	"..DcCCCCcD......",
	"...DcCCcD.......",
	"....DDDD........",
	"................",
	"................",
}

// loadPixelItems builds the code-drawn item sprites (called from loadSprites).
func loadPixelItems() (coffee, pill *ebiten.Image) {
	return buildPixelSprite(coffeeArt, pixelPal), buildPixelSprite(pillArt, pixelPal)
}
