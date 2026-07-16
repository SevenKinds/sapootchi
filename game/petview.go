package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/ui"
)

// drawPetAndState renders the pet with its current state pill — the same view
// on every page that shows the pet (Home, Inventory), so he is always doing
// the same thing everywhere: asleep is asleep, energized is energized.
// Returns false if the pet is away (caller decides what to show instead).
func drawPetAndState(g *Game, screen *ebiten.Image, blobCy, pillY float64) bool {
	pet := g.Pet
	if pet.Away {
		ui.DrawTextCenter(screen, pet.Name+" ran away!", ScreenW/2, blobCy-20, 22, ui.Text, true)
		ui.DrawTextCenter(screen, "Leave food out on the Home tab.", ScreenW/2, blobCy+16, 13, ui.TextDim, false)
		return false
	}

	g.DrawBlob(screen, ScreenW/2, blobCy)
	switch {
	case pet.Asleep:
		drawStatePill(screen, "Zzz... asleep until 50% energy", ui.Secondary, pillY)
	case pet.Energized():
		drawStatePill(screen, "Bursting with energy — let's play!", ui.Energy, pillY)
	default:
		drawStatePill(screen, "Mood: "+pet.Mood().String(), ui.Panel, pillY)
	}
	return true
}

func drawStatePill(screen *ebiten.Image, msg string, clr color.Color, y float64) {
	w := ui.TextWidth(msg, 12, true) + 28
	x := (ScreenW - w) / 2
	ui.FillRoundRect(screen, float32(x), float32(y), float32(w), 24, 12, clr)
	ui.DrawTextCenter(screen, msg, ScreenW/2, y+6, 12, ui.Text, true)
}
