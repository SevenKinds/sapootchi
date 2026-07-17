package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/simulation"
	"sapootchi/ui"
)

// drawPetAndState renders the pet (centered on cx) with its current state
// pill — the same view on every page that shows the pet (Home, Inventory), so
// he is always doing the same thing everywhere.
// Returns false if the pet is away (caller decides what to show instead).
func drawPetAndState(g *Game, screen *ebiten.Image, cx, blobCy, pillY float64) bool {
	pet := g.Pet
	if pet.Away {
		ui.DrawTextCenter(screen, pet.Name+" ran away!", ScreenW/2, blobCy-20, 22, ui.Text, true)
		ui.DrawTextCenter(screen, "Leave food out on the Home tab.", ScreenW/2, blobCy+16, 13, ui.TextDim, false)
		return false
	}

	g.DrawBlob(screen, cx, blobCy)
	switch {
	case pet.Asleep:
		drawStatePill(screen, "Zzz... asleep until rested", ui.Secondary, cx, pillY)
	case pet.Energized():
		drawStatePill(screen, "Bursting with energy!", ui.Energy, cx, pillY)
	default:
		drawStatePill(screen, "Mood: "+pet.Mood().String(), ui.Panel, cx, pillY)
	}
	return true
}

func drawStatePill(screen *ebiten.Image, msg string, clr color.Color, cx, y float64) {
	w := ui.TextWidth(msg, 12, true) + 28
	ui.FillRoundRect(screen, float32(cx-w/2), float32(y), float32(w), 24, 12, clr)
	ui.DrawTextCenter(screen, msg, cx, y+6, 12, ui.Text, true)
}

// statCells: the four visible stats as glyph + value, shared by the slim
// vertical column (Home, Inventory).
var statCells = []struct {
	glyph rune
	fixed color.Color // nil = color by value
	get   func(s simulation.Stats) float64
}{
	{'', nil, func(s simulation.Stats) float64 { return s.Happiness }},
	{'', nil, func(s simulation.Stats) float64 { return s.Hunger }},
	{'', nil, func(s simulation.Stats) float64 { return s.Hygiene }},
	{'', ui.Energy, func(s simulation.Stats) float64 { return s.Energy }},
}

// drawStatColumn renders the slim stats stacked vertically at (x, y): glyph +
// percentage on a line, a mini bar underneath. ~46px per cell, barW wide.
func drawStatColumn(g *Game, screen *ebiten.Image, x, y, barW float64) {
	for i, c := range statCells {
		cy := y + float64(i)*46
		v := c.get(g.Pet.Stats)

		clr := c.fixed
		if clr == nil {
			clr = statColorFor(v)
		}
		ui.DrawGlyph(screen, c.glyph, x+10, cy+8, 13, clr)
		ui.DrawTextBold(screen, ui.Itoa(int(v+0.5))+"%", x+24, cy, 12, ui.Text)
		ui.FillRoundRect(screen, float32(x), float32(cy+20), float32(barW), 6, 3, ui.Track)
		if fw := barW * v / 100; fw > 6 {
			ui.FillRoundRect(screen, float32(x), float32(cy+20), float32(fw), 6, 3, clr)
		}
	}
}

// statColorFor mirrors the StatBar coloring.
func statColorFor(v float64) color.Color {
	switch {
	case v < 25:
		return ui.Bad
	case v < 50:
		return ui.Warn
	default:
		return ui.Good
	}
}
