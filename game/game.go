package game

import (
	_ "image/png" // register PNG decoder for the embedded sprites
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

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
// live here. Storage is platform-specific (file on desktop, localStorage on
// WASM — see storage_*.go).
type Game struct {
	// Pet is the ACTIVE pet — always == Pets[Active]. All pets keep living
	// (decaying, sleeping, regenerating) whether active or not.
	Pet      *simulation.Pet
	Pets     []*simulation.Pet
	Active   int
	Settings Settings
	Rng      *rand.Rand
	Sprites  *spriteBank

	scenes []Scene
	tick   int

	lastPerfectYearDay int // day-of-year the Perfect Care bonus was last given
}

// MaxPets caps the roster (tadpoles from the shop add pets up to this).
const MaxPets = 3

// New builds the game, loading the saved roster or rescuing a first pet.
func New() *Game {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	var pets []*simulation.Pet
	var active int
	var settings Settings
	if data, err := readSave(); err == nil {
		if ps, a, s, err := decodeSave(data); err == nil {
			pets, active, settings = ps, a, s
		}
	}
	if len(pets) == 0 {
		pets = []*simulation.Pet{simulation.NewRandomPet("Blobby", time.Now(), rng)}
		active = 0
	}

	g := &Game{
		Pet:      pets[active],
		Pets:     pets,
		Active:   active,
		Settings: settings,
		Rng:      rng,
		Sprites:  loadSprites(),
	}
	g.Push(NewMainScene())
	return g
}

// SwitchPet makes the next pet in the roster active and returns it.
func (g *Game) SwitchPet() *simulation.Pet {
	g.Active = (g.Active + 1) % len(g.Pets)
	g.Pet = g.Pets[g.Active]
	g.Save()
	return g.Pet
}

// AddPet appends a new pet and makes it active.
func (g *Game) AddPet(p *simulation.Pet) {
	g.Pets = append(g.Pets, p)
	g.Active = len(g.Pets) - 1
	g.Pet = p
	g.Save()
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
	// EVERY pet keeps living, not just the active one.
	now := time.Now()
	for _, p := range g.Pets {
		p.Update(now, g.Rng)
	}

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
// the current phase (baby = 60%). The pose follows state: equipped skin, or on
// the classic skin the current mood (Zz pose while asleep). Design-space coords.
func (g *Game) DrawBlob(screen *ebiten.Image, cx, cy float64) {
	img := g.petSprite()
	bw := float64(img.Bounds().Dx())
	bh := float64(img.Bounds().Dy())
	const target = 190.0 // full-size display width in design px
	scale := (target / bw) * g.Pet.Phase.RenderScale()
	bob := math.Sin(float64(g.tick)/30.0) * 4
	if g.Pet.Asleep {
		bob = math.Sin(float64(g.tick)/55.0) * 2 // slow breathing
	}

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
	screen.DrawImage(img, op)
}

// Save writes the roster + settings to platform storage (best-effort).
func (g *Game) Save() {
	data, err := encodeSave(g.Pets, g.Active, g.Settings)
	if err != nil {
		return
	}
	_ = writeSave(data)
}
