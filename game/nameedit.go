package game

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"sapootchi/ui"
)

// NameEditModal lets the player rename the active pet: a text field fed by the
// hardware keyboard (desktop + web), with on-screen OK/Cancel for taps. Enter
// confirms, Esc cancels.
type NameEditModal struct {
	blurUnder
	text string
	tick int
}

const nameMaxLen = 14

func NewNameEditModal(under Scene, current string) *NameEditModal {
	return &NameEditModal{blurUnder: blurUnder{under: under}, text: current}
}

// Card geometry.
const (
	neW = 300.0
	neH = 250.0
)

func (m *NameEditModal) buttons() (ok, cancel ui.Button) {
	x := (ScreenW - neW) / 2
	y := (ScreenH - neH) / 2
	bw := (neW - 3*16) / 2
	ok = ui.Button{X: x + 16*2 + bw, Y: y + neH - 60, W: bw, H: 44, Label: "Save"}
	cancel = ui.Button{X: x + 16, Y: y + neH - 60, W: bw, H: 44, Label: "Cancel", Secondary: true}
	return
}

func (m *NameEditModal) Update(g *Game) error {
	m.tick++
	if m.tick < 8 {
		return nil // grace: don't eat the opening tap
	}

	// Typed characters (printable ASCII only, capped).
	for _, r := range ebiten.AppendInputChars(nil) {
		if r >= 32 && r < 127 && len(m.text) < nameMaxLen {
			m.text += string(r)
		}
	}
	// Backspace: once on press, then repeat while held.
	if d := inpututil.KeyPressDuration(ebiten.KeyBackspace); (d == 1 || (d > 25 && d%3 == 0)) && len(m.text) > 0 {
		m.text = m.text[:len(m.text)-1]
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter) {
		m.commit(g)
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.Pop()
		return nil
	}

	ok, cancel := m.buttons()
	switch {
	case ok.Clicked():
		m.commit(g)
	case cancel.Clicked():
		g.Pop()
	}
	return nil
}

func (m *NameEditModal) commit(g *Game) {
	if name := strings.TrimSpace(m.text); name != "" {
		g.Pet.Name = name
		g.Save()
	}
	g.Pop()
}

func (m *NameEditModal) Draw(g *Game, screen *ebiten.Image) {
	m.blurUnder.draw(g, screen)

	x := (ScreenW - neW) / 2
	y := (ScreenH - neH) / 2
	drawModalCard(screen, x, y, neW, neH)

	ui.DrawTextCenter(screen, "Name your pet", ScreenW/2, y+26, 18, ui.Text, true)

	// Text field.
	fx, fy, fw, fh := x+24, y+72, neW-48, 46.0
	ui.FillRoundRect(screen, float32(fx), float32(fy), float32(fw), float32(fh), 10, ui.Panel)
	shown := m.text
	tw := ui.TextWidth(shown, 18, true)
	tx := fx + 16
	ui.DrawTextBold(screen, shown, tx, fy+fh/2-11, 18, ui.Text)
	if (m.tick/30)%2 == 0 { // blinking caret
		cx := tx + tw + 2
		ui.FillRoundRect(screen, float32(cx), float32(fy+11), 2, float32(fh-22), 1, ui.Accent)
	}

	ui.DrawTextCenter(screen, "type a name · Enter to save", ScreenW/2, y+134, 11, ui.TextDim, false)

	ok, cancel := m.buttons()
	ok.Draw(screen, strings.TrimSpace(m.text) != "")
	cancel.Draw(screen, true)
}
