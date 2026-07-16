package game

import (
	"bytes"
	"image"
	_ "image/png" // register PNG decoder for the embedded blob sprite
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/assets"
	"sapootchi/simulation"
	"sapootchi/ui"
)

// Logical screen size (portrait, mobile-ish) in design-space units. Scenes
// author against this; ui.Scale raises the actual framebuffer resolution.
const (
	ScreenW = 360
	ScreenH = 640
)

// Game is the Ebiten root: it owns the pet, settings, a scene stack, and
// persistence. It is a *view + input* layer over simulation — no game rules
// live here.
type Game struct {
	Pet      *simulation.Pet
	Settings Settings
	Rng      *rand.Rand
	Blob     *ebiten.Image

	scenes   []Scene
	savePath string
	tick     int

	lastPerfectYearDay int // day-of-year the Perfect Care bonus was last given
}

// New builds the game, loading a saved pet if one exists or hatching a new one.
func New() *Game {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	savePath := savefilePath()
	var pet *simulation.Pet
	var settings Settings
	if data, err := os.ReadFile(savePath); err == nil {
		if p, s, err := decodeSave(data); err == nil {
			pet, settings = p, s
		}
	}
	if pet == nil {
		pet = simulation.NewRandomPet("Blobby", time.Now(), rng)
	}

	g := &Game{
		Pet:      pet,
		Settings: settings,
		Rng:      rng,
		Blob:     decodeImage(assets.BlobPNG),
		savePath: savePath,
	}
	g.Push(NewMainScene())
	return g
}

// --- scene stack ---

func (g *Game) Push(s Scene) { g.scenes = append(g.scenes, s) }

func (g *Game) Pop() {
	if len(g.scenes) > 1 {
		g.scenes = g.scenes[:len(g.scenes)-1]
	}
}

func (g *Game) current() Scene { return g.scenes[len(g.scenes)-1] }

// --- ebiten.Game ---

func (g *Game) Update() error {
	g.tick++
	ui.UpdateInput()
	// Real-time decay every tick (works off wall clock, not frame count).
	g.Pet.Update(time.Now(), g.Rng)

	if err := g.current().Update(g); err != nil {
		return err
	}

	// Autosave every ~5s.
	if g.tick%300 == 0 {
		g.Save()
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ui.BackgroundGradient(screen)
	g.current().Draw(g, screen)
}

// Layout renders into a framebuffer scaled up from the 360x640 design space, so
// text and shapes are drawn at higher resolution. Ebiten downscales it to the
// window, which can be any size (see main).
func (g *Game) Layout(int, int) (int, int) {
	return int(ScreenW * ui.Scale), int(ScreenH * ui.Scale)
}

// --- helpers ---

// DrawBlob draws the pet centered at (cx,cy) with a gentle idle bob, scaled for
// the current phase (baby = 60%). Coordinates are design-space.
func (g *Game) DrawBlob(screen *ebiten.Image, cx, cy float64) {
	bw := float64(g.Blob.Bounds().Dx())
	bh := float64(g.Blob.Bounds().Dy())
	const target = 190.0 // full-size display width in design px
	scale := (target / bw) * g.Pet.Phase.RenderScale()
	bob := math.Sin(float64(g.tick)/30.0) * 4

	// Soft shadow at the feet (shrinks with the pet). ui.FillRoundRect scales
	// its own coordinates, so pass design-space values.
	shW := bw * scale * 0.55
	const shH = 16.0
	ui.FillRoundRect(screen, float32(cx-shW/2), float32(cy+bh*scale/2-shH),
		float32(shW), shH, shH/2, ui.Shadow)

	// The sprite is drawn directly, so apply the render scale here.
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale*ui.Scale, scale*ui.Scale)
	op.GeoM.Translate((cx-bw*scale/2)*ui.Scale, (cy-bh*scale/2+bob)*ui.Scale)
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(g.Blob, op)
}

// Save writes the pet + settings to disk (best-effort).
func (g *Game) Save() {
	data, err := encodeSave(g.Pet, g.Settings)
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(g.savePath), 0o755)
	_ = os.WriteFile(g.savePath, data, 0o644)
}

func savefilePath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "sapootchi_save.json"
	}
	return filepath.Join(dir, "sapootchi", "save.json")
}

func decodeImage(b []byte) *ebiten.Image {
	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		log.Fatalf("decode sprite: %v", err)
	}
	return ebiten.NewImageFromImage(img)
}
