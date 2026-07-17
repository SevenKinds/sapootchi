package game

import (
	"bytes"
	"image"
	"io/fs"
	"log"
	"path"
	"sort"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/assets"
	"sapootchi/simulation"
	"sapootchi/ui"
)

// spriteBank is all decoded pet art: the classic blob, the mood poses, and the
// themed skins from the brand kit.
type spriteBank struct {
	Blob       *ebiten.Image                         // classic/neutral
	Moods      map[simulation.Mood]*ebiten.Image     // pose per mood (default skin only)
	Asleep     *ebiten.Image                         // Zz pose for the forced nap
	React      map[string]*ebiten.Image              // reaction poses: "wink", "hearts", "angry"
	Skins      map[string]*ebiten.Image              // skin name -> full art
	SkinNames  []string                              // sorted, for UI
	Anims      map[string][]*ebiten.Image            // reaction animations (modal player)
	Clouds     []*ebiten.Image                       // background clouds (Home, Climber)
	Coin       []*ebiten.Image                       // 12-frame spinning coin (16x16 px art)
	Platforms  *ebiten.Image                         // brackeys platform sheet (Climber)
	RiverRocks [][]*ebiten.Image                     // animated rock variants (River)
	Foam       []*ebiten.Image                       // animated foam patches (River water texture)
	Items      map[simulation.FoodKind]*ebiten.Image // pixel item sprites (apple/steak/berries)
	Fruits     []*ebiten.Image                       // brackeys fruit sprites (catch-food)
	Meat       *ebiten.Image                         // steak sprite
	Bag        *ebiten.Image                         // money-bag catcher (catch-food)
}

// skinDisplay names the EMOJIS 3D poses used as skins.
var skinDisplay = map[string]string{
	"sapo_01": "Wide Eyes", "sapo_02": "Wink", "sapo_03": "Side Eye",
	"sapo_05": "Unimpressed", "sapo_06": "Cheerful", "sapo_07": "Bright",
	"sapo_08": "Deadpan", "sapo_09": "Gloomy", "sapo_10": "Snoozy",
	"sapo_11": "Curious", "sapo_12": "Shy", "sapo_13": "Grumpy",
	"sapo_14": "Calm", "sapo_15": "Specs", "sapo_16": "Shades",
	"sapo_17": "In Love", "sapo_18": "Starstruck", "sapo_19": "Pride",
}

// Mood pose mapping, chosen by eye from the brand's EMOJIS 3D set.
var moodFiles = map[simulation.Mood]string{
	simulation.MoodHappy:   "sapo_06.png", // cheerful soft eyes
	simulation.MoodHungry:  "sapo_09.png", // droopy, miserable
	simulation.MoodExcited: "sapo_18.png", // star eyes
	simulation.MoodBored:   "sapo_05.png", // heavy lids
	simulation.MoodSleepy:  "sapo_08.png", // half-shut deadpan
	simulation.MoodCurious: "sapo_03.png", // side-glance
	simulation.MoodLonely:  "sapo_12.png", // shy, looking down
}

const asleepFile = "sapo_10.png" // Zz

// Reaction poses banked for events (feeding, petting, ...).
var reactFiles = map[string]string{
	"wink":   "sapo_02.png",
	"hearts": "sapo_17.png",
	"angry":  "sapo_13.png",
}

func loadSprites() *spriteBank {
	b := &spriteBank{
		Blob:  decodeImage(assets.BlobPNG),
		Moods: map[simulation.Mood]*ebiten.Image{},
		React: map[string]*ebiten.Image{},
		Skins: map[string]*ebiten.Image{},
	}

	for mood, file := range moodFiles {
		b.Moods[mood] = loadFS("sprites/moods/" + file)
	}
	b.Asleep = loadFS("sprites/moods/" + asleepFile)
	for name, file := range reactFiles {
		b.React[name] = loadFS("sprites/moods/" + file)
	}

	// Skins are the EMOJIS 3D poses (uniform sizes, unlike the old themed set).
	entries, err := fs.ReadDir(assets.Sprites, "sprites/moods")
	if err == nil {
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".png") {
				continue
			}
			name := strings.TrimSuffix(e.Name(), ".png")
			b.Skins[name] = loadFS(path.Join("sprites/moods", e.Name()))
			b.SkinNames = append(b.SkinNames, name)
		}
		sort.Strings(b.SkinNames)
	}

	// Background clouds.
	if entries, err := fs.ReadDir(assets.Sprites, "sprites/clouds"); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".png") {
				b.Clouds = append(b.Clouds, loadFS(path.Join("sprites/clouds", e.Name())))
			}
		}
	}

	// Spinning coin: a horizontal strip of square frames.
	if sheet := loadFS("sprites/icons/coin.png"); sheet != nil {
		size := sheet.Bounds().Dy()
		for x := 0; x+size <= sheet.Bounds().Dx(); x += size {
			b.Coin = append(b.Coin,
				sheet.SubImage(image.Rect(x, 0, x+size, size)).(*ebiten.Image))
		}
	}

	// Mini-game art: platform sheet + animated river rocks (square frames in
	// horizontal strips).
	b.Platforms = loadFS("sprites/games/platforms.png")
	for _, name := range []string{"rock_1.png", "rock_2.png", "rock_3.png", "rock_4.png"} {
		sheet := loadFS("sprites/games/" + name)
		size := sheet.Bounds().Dy()
		var frames []*ebiten.Image
		for x := 0; x+size <= sheet.Bounds().Dx(); x += size {
			frames = append(frames, sheet.SubImage(image.Rect(x, 0, x+size, size)).(*ebiten.Image))
		}
		b.RiverRocks = append(b.RiverRocks, frames)
	}
	if sheet := loadFS("sprites/games/foam.png"); sheet != nil {
		size := sheet.Bounds().Dy()
		for x := 0; x+size <= sheet.Bounds().Dx(); x += size {
			b.Foam = append(b.Foam, sheet.SubImage(image.Rect(x, 0, x+size, size)).(*ebiten.Image))
		}
	}

	// Item sprites: brackeys fruit grid (16px, 3 cols x 4 color rows) + Tiny
	// Swords meat & money bag.
	fruit := loadFS("sprites/items/fruit.png")
	for row := 0; row < 4; row++ {
		for col := 0; col < 3; col++ {
			b.Fruits = append(b.Fruits,
				fruit.SubImage(image.Rect(col*16, row*16, col*16+16, row*16+16)).(*ebiten.Image))
		}
	}
	b.Meat = loadFS("sprites/items/meat.png")
	b.Bag = loadFS("sprites/items/bag.png")
	b.Items = map[simulation.FoodKind]*ebiten.Image{
		simulation.FoodApple:    b.Fruits[9], // red apple (row 3, col 0)
		simulation.FoodSandwich: b.Meat,
		simulation.FoodCake:     b.Fruits[8], // purple grapes (row 2, col 2)
	}
	coffee, pill := loadPixelItems()
	b.Items[simulation.FoodCoffee] = coffee
	b.Items[simulation.FoodEnergyPill] = pill
	itemSprites = b.Items // package-level for drawItemIcon

	// Animations: each subdir of sprites/anims is a frame sequence.
	b.Anims = map[string][]*ebiten.Image{}
	if dirs, err := fs.ReadDir(assets.Sprites, "sprites/anims"); err == nil {
		for _, d := range dirs {
			if !d.IsDir() {
				continue
			}
			frameDir := path.Join("sprites/anims", d.Name())
			frames, err := fs.ReadDir(assets.Sprites, frameDir)
			if err != nil {
				continue
			}
			var seq []*ebiten.Image
			for _, f := range frames { // ReadDir returns sorted names
				if strings.HasSuffix(f.Name(), ".png") {
					seq = append(seq, loadFS(path.Join(frameDir, f.Name())))
				}
			}
			if len(seq) > 0 {
				b.Anims[d.Name()] = seq
			}
		}
	}
	return b
}

func loadFS(p string) *ebiten.Image {
	data, err := fs.ReadFile(assets.Sprites, p)
	if err != nil {
		log.Fatalf("read sprite %s: %v", p, err)
	}
	return decodeImage(data)
}

func decodeImage(b []byte) *ebiten.Image {
	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		log.Fatalf("decode sprite: %v", err)
	}
	return ebiten.NewImageFromImage(img)
}

// petSprite picks what the ACTIVE pet looks like right now: its own equipped
// skin is a full look; on the classic skin the pose follows state/mood, with
// transient reactions (hearts on petting) on top.
func (g *Game) petSprite() *ebiten.Image {
	if img, ok := g.Sprites.Skins[g.Pet.Skin]; ok {
		return img
	}
	if g.Pet.Asleep {
		return g.Sprites.Asleep
	}
	if g.tick < g.reactUntil && g.reactImg != nil {
		return g.reactImg
	}
	if img, ok := g.Sprites.Moods[g.Pet.Mood()]; ok {
		return img
	}
	return g.Sprites.Blob
}

// ShowReaction flashes a reaction pose ("hearts", "wink", "angry") on the
// classic-skin pet for the given ticks.
func (g *Game) ShowReaction(name string, ticks int) {
	if img, ok := g.Sprites.React[name]; ok {
		g.reactImg = img
		g.reactUntil = g.tick + ticks
	}
}

// CoinFrame returns the current frame of the spinning coin (~10fps).
func (g *Game) CoinFrame() *ebiten.Image {
	if len(g.Sprites.Coin) == 0 {
		return nil
	}
	return g.Sprites.Coin[(g.tick/6)%len(g.Sprites.Coin)]
}

// DrawCoin draws the spinning coin at (x, y) with the given design-px size.
func (g *Game) DrawCoin(dst *ebiten.Image, x, y, size float64) {
	if c := g.CoinFrame(); c != nil {
		g.drawCoinImg(dst, c, x, y, size)
		return
	}
	// Fallback: static gold dot.
	ui.FillCircle(dst, float32(x+size/2), float32(y+size/2), float32(size/2), ui.Gold)
}

func (g *Game) drawCoinImg(dst *ebiten.Image, c *ebiten.Image, x, y, size float64) {
	f := size / float64(c.Bounds().Dx())
	ui.DrawImageNearest(dst, c, x, y, f, 1)
}

// baseSprite is the active pet's look ignoring transient mood — used by
// mini-games.
func (g *Game) baseSprite() *ebiten.Image {
	if img, ok := g.Sprites.Skins[g.Pet.Skin]; ok {
		return img
	}
	return g.Sprites.Blob
}

// OwnsSkin reports whether a look is unlocked ("" = classic, always owned).
func (g *Game) OwnsSkin(name string) bool {
	if name == "" {
		return true
	}
	for _, n := range g.Settings.OwnedSkins {
		if n == name {
			return true
		}
	}
	return false
}

// UnownedSkins lists locked looks in reveal order (alphabetical); the shop
// shows the first few.
func (g *Game) UnownedSkins() []string {
	var out []string
	for _, n := range g.Sprites.SkinNames {
		if !g.OwnsSkin(n) {
			out = append(out, n)
		}
	}
	return out
}
