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
)

// spriteBank is all decoded pet art: the classic blob, the mood poses, and the
// themed skins from the brand kit.
type spriteBank struct {
	Blob      *ebiten.Image                     // classic/neutral
	Moods     map[simulation.Mood]*ebiten.Image // pose per mood (default skin only)
	Asleep    *ebiten.Image                     // Zz pose for the forced nap
	React     map[string]*ebiten.Image          // reaction poses: "wink", "hearts", "angry"
	Skins     map[string]*ebiten.Image          // skin name -> full art
	SkinNames []string                          // sorted, for UI
	Anims     map[string][]*ebiten.Image        // reaction animations (modal player)
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

	// Skins: every png in sprites/skins.
	entries, err := fs.ReadDir(assets.Sprites, "sprites/skins")
	if err == nil {
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".png") {
				continue
			}
			name := strings.TrimSuffix(e.Name(), ".png")
			b.Skins[name] = loadFS(path.Join("sprites/skins", e.Name()))
			b.SkinNames = append(b.SkinNames, name)
		}
		sort.Strings(b.SkinNames)
	}

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
// skin is a full look; on the classic skin the pose follows state/mood.
func (g *Game) petSprite() *ebiten.Image {
	if img, ok := g.Sprites.Skins[g.Pet.Skin]; ok {
		return img
	}
	if g.Pet.Asleep {
		return g.Sprites.Asleep
	}
	if img, ok := g.Sprites.Moods[g.Pet.Mood()]; ok {
		return img
	}
	return g.Sprites.Blob
}

// baseSprite is the active pet's look ignoring transient mood — used by
// mini-games.
func (g *Game) baseSprite() *ebiten.Image {
	if img, ok := g.Sprites.Skins[g.Pet.Skin]; ok {
		return img
	}
	return g.Sprites.Blob
}
