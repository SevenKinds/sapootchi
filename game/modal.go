package game

import (
	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/ui"
)

// blurUnder renders the scene beneath a modal softly blurred (downscale then
// upscale with linear filtering). Shared by every modal.
type blurUnder struct {
	under    Scene
	underBuf *ebiten.Image
	smallBuf *ebiten.Image
}

// blurFactor: the downscale divisor — bigger = softer.
const blurFactor = 7

func (m *blurUnder) draw(g *Game, screen *ebiten.Image) {
	fw, fh := screen.Bounds().Dx(), screen.Bounds().Dy()
	if m.underBuf == nil || m.underBuf.Bounds().Dx() != fw || m.underBuf.Bounds().Dy() != fh {
		m.underBuf = ebiten.NewImage(fw, fh)
		m.smallBuf = ebiten.NewImage(fw/blurFactor+1, fh/blurFactor+1)
	}

	m.underBuf.Clear()
	ui.BackgroundGradient(m.underBuf)
	m.under.Draw(g, m.underBuf)

	m.smallBuf.Clear()
	down := &ebiten.DrawImageOptions{}
	down.GeoM.Scale(1.0/blurFactor, 1.0/blurFactor)
	down.Filter = ebiten.FilterLinear
	m.smallBuf.DrawImage(m.underBuf, down)

	up := &ebiten.DrawImageOptions{}
	up.GeoM.Scale(blurFactor, blurFactor)
	up.Filter = ebiten.FilterLinear
	screen.DrawImage(m.smallBuf, up)
}

// drawModalCard paints a rounded card blended into the page gradient (rounded
// caps in their local color, square bands between) with a soft shadow.
func drawModalCard(screen *ebiten.Image, x, y, w, h float64) {
	ui.FillRoundRect(screen, float32(x), float32(y+5), float32(w), float32(h), 20, ui.Shadow)
	ui.FillRoundRect(screen, float32(x), float32(y), float32(w), 40, 20, ui.BGColorAt(y+10, ScreenH))
	ui.FillRoundRect(screen, float32(x), float32(y+h-40), float32(w), 40, 20, ui.BGColorAt(y+h-10, ScreenH))
	const bands = 14
	inner := h - 40
	for i := 0; i < bands; i++ {
		by := y + 20 + inner*float64(i)/bands
		ui.FillRoundRect(screen, float32(x), float32(by), float32(w), float32(inner/bands+1), 0,
			ui.BGColorAt(by, ScreenH))
	}
}

// ConfirmModal asks before a consequential action: preview, optional price,
// OK/Cancel. price < 0 hides the coin row (non-purchase confirms like Rest).
type ConfirmModal struct {
	blurUnder
	title    string
	subtitle string
	img      *ebiten.Image
	price    int
	okLabel  string
	onOK     func(g *Game)
	tick     int
}

func NewConfirmModal(title, subtitle string, img *ebiten.Image, price int, under Scene, onBuy func(g *Game)) *ConfirmModal {
	return &ConfirmModal{
		blurUnder: blurUnder{under: under},
		title:     title, subtitle: subtitle,
		img: img, price: price, okLabel: "Buy", onOK: onBuy,
	}
}

// NewConfirmAction is a purchase-free confirm (e.g. tucking the pet in).
func NewConfirmAction(title, subtitle string, img *ebiten.Image, okLabel string, under Scene, onOK func(g *Game)) *ConfirmModal {
	return &ConfirmModal{
		blurUnder: blurUnder{under: under},
		title:     title, subtitle: subtitle,
		img: img, price: -1, okLabel: okLabel, onOK: onOK,
	}
}

// Card geometry.
const (
	cfW = 280.0
	cfH = 330.0
)

func (m *ConfirmModal) buttons() (buy, cancel ui.Button) {
	x := (ScreenW - cfW) / 2
	y := (ScreenH - cfH) / 2
	bw := (cfW - 3*16) / 2
	buy = ui.Button{X: x + 16*2 + bw, Y: y + cfH - 60, W: bw, H: 44, Label: m.okLabel}
	cancel = ui.Button{X: x + 16, Y: y + cfH - 60, W: bw, H: 44, Label: "Cancel", Secondary: true}
	return
}

func (m *ConfirmModal) Update(g *Game) error {
	m.tick++
	if m.tick < 12 {
		return nil // grace: don't eat the opening tap
	}
	buy, cancel := m.buttons()
	switch {
	case buy.Clicked():
		m.onOK(g)
		g.Pop()
	case cancel.Clicked():
		g.Pop()
	}
	return nil
}

func (m *ConfirmModal) Draw(g *Game, screen *ebiten.Image) {
	m.blurUnder.draw(g, screen)

	x := (ScreenW - cfW) / 2
	y := (ScreenH - cfH) / 2
	drawModalCard(screen, x, y, cfW, cfH)

	ui.DrawTextCenter(screen, m.title, ScreenW/2, y+22, 17, ui.Text, true)
	if m.img != nil {
		ui.DrawImageFit(screen, m.img, x+50, y+52, cfW-100, 130)
	}
	ui.DrawTextCenter(screen, m.subtitle, ScreenW/2, y+196, 12, ui.TextDim, false)

	// Price row with the spinning coin (purchases only).
	if m.price >= 0 {
		priceStr := ui.Itoa(m.price)
		pw := ui.TextWidth(priceStr, 18, true) + 26
		g.DrawCoin(screen, ScreenW/2-pw/2, y+226, 20)
		ui.DrawTextBold(screen, priceStr, ScreenW/2-pw/2+26, y+227, 18, ui.Gold)
	}

	buy, cancel := m.buttons()
	buy.Draw(screen, m.price < 0 || g.Pet.Coins >= m.price)
	cancel.Draw(screen, true)
}
