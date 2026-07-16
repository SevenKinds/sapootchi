// Command assetprep converts brand source art (artpacks/) into game-ready
// sprites (assets/sprites/). Reproducible ingestion: run it again whenever the
// brand kit changes.
//
//	go run ./cmd/assetprep
//
// It does three things per image:
//  1. White removal (TIFs only): flood-fill from the canvas edges, so the
//     white BACKGROUND becomes transparent but white DETAILS inside the
//     mascot (the eyes!) survive.
//  2. Trim to the alpha bounding box (plus padding) so sprites are tightly
//     framed regardless of source canvas layout.
//  3. Downscale to a game-friendly size (max 384px, CatmullRom).
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/image/draw"
	"golang.org/x/image/tiff"
)

const (
	maxDim  = 384 // output max dimension in px
	trimPad = 24  // px of padding kept around the alpha bbox (at source scale)
	whiteT  = 238 // channel threshold: >= this on r,g,b counts as "white"
)

// Curated animation set (modal reactions). All 16 would cost ~16+ MB embedded;
// these five cover the states the player actually sees.
var anims = map[string]string{
	"sleepy":   "artpacks/brand/EMOJIS ANIMADOS/SAPO_SLEEPY.mp4",
	"triste":   "artpacks/brand/EMOJIS ANIMADOS/SAPO_TRISTE.mp4",
	"wink":     "artpacks/brand/EMOJIS ANIMADOS/SAPO_WINK.mp4",
	"coracoes": "artpacks/brand/EMOJIS ANIMADOS/SAPO_CORACOES.mp4",
	"estrelas": "artpacks/brand/EMOJIS ANIMADOS/SAPO_ESTRELAS.mp4",
}

const (
	animFPS   = 10
	animWidth = 320
)

func main() {
	moods := convertDir("artpacks/brand/EMOJIS 3D", "assets/sprites/moods", ".tif", true)
	skins := convertDir("artpacks/brand/EMOJIS ESPECIAIS", "assets/sprites/skins", ".png", false)
	nAnims := convertAnims()
	fmt.Printf("done: %d mood sprites, %d skins, %d animations\n", moods, skins, nAnims)
}

// convertAnims extracts curated mp4s to frame sequences: ffmpeg samples at
// animFPS, then each frame gets white removal, and ALL frames are cropped to
// the UNION alpha bbox (uniform crop — per-frame trims would make the loop
// jitter).
func convertAnims() int {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		fmt.Println("ffmpeg not found — skipping animations")
		return 0
	}
	n := 0
	for name, src := range anims {
		if err := convertAnim(name, src); err != nil {
			log.Fatalf("anim %s: %v", name, err)
		}
		fmt.Printf("  %s -> assets/sprites/anims/%s\n", src, name)
		n++
	}
	return n
}

func convertAnim(name, src string) error {
	tmp, err := os.MkdirTemp("", "assetprep-"+name)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	cmd := exec.Command("ffmpeg", "-y", "-loglevel", "error", "-i", src,
		"-vf", fmt.Sprintf("fps=%d,scale=%d:-1", animFPS, animWidth),
		filepath.Join(tmp, "%03d.png"))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg: %v: %s", err, out)
	}

	names, err := filepath.Glob(filepath.Join(tmp, "*.png"))
	if err != nil {
		return err
	}
	sort.Strings(names)

	// Pass 1: decode + white removal + track the union bbox.
	var frames []*image.RGBA
	minX, minY, maxX, maxY := 1<<30, 1<<30, -1, -1
	for _, fn := range names {
		f, err := os.Open(fn)
		if err != nil {
			return err
		}
		img, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			return err
		}
		rgba := toRGBA(img)
		floodWhite(rgba)
		frames = append(frames, rgba)
		b := bbox(rgba)
		if b.Min.X < minX {
			minX = b.Min.X
		}
		if b.Min.Y < minY {
			minY = b.Min.Y
		}
		if b.Max.X > maxX {
			maxX = b.Max.X
		}
		if b.Max.Y > maxY {
			maxY = b.Max.Y
		}
	}
	if maxX < 0 {
		return fmt.Errorf("no visible pixels")
	}

	// Pass 2: uniform crop + save.
	dstDir := filepath.Join("assets/sprites/anims", name)
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return err
	}
	crop := image.Rect(minX, minY, maxX+1, maxY+1)
	for i, fr := range frames {
		dst := image.NewRGBA(image.Rect(0, 0, crop.Dx(), crop.Dy()))
		draw.Draw(dst, dst.Bounds(), fr, crop.Min, draw.Src)
		out, err := os.Create(filepath.Join(dstDir, fmt.Sprintf("%03d.png", i)))
		if err != nil {
			return err
		}
		if err := png.Encode(out, dst); err != nil {
			out.Close()
			return err
		}
		out.Close()
	}
	return nil
}

// bbox returns the alpha bounding box of img.
func bbox(img *image.RGBA) image.Rectangle {
	b := img.Bounds()
	minX, minY, maxX, maxY := b.Max.X, b.Max.Y, b.Min.X-1, b.Min.Y-1
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.RGBAAt(x, y).A > 8 {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}
	return image.Rect(minX, minY, maxX+1, maxY+1)
}

func convertDir(srcDir, dstDir, ext string, removeWhite bool) int {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		log.Fatalf("read %s: %v", srcDir, err)
	}
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		log.Fatal(err)
	}
	n := 0
	for _, e := range entries {
		if e.IsDir() || !strings.EqualFold(filepath.Ext(e.Name()), ext) {
			continue
		}
		src := filepath.Join(srcDir, e.Name())
		base := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		base = strings.ToLower(strings.ReplaceAll(base, " ", "_"))
		dst := filepath.Join(dstDir, base+".png")
		if err := convert(src, dst, removeWhite); err != nil {
			log.Fatalf("%s: %v", src, err)
		}
		fmt.Printf("  %s -> %s\n", src, dst)
		n++
	}
	return n
}

func convert(srcPath, dstPath string, removeWhite bool) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()

	var img image.Image
	if strings.EqualFold(filepath.Ext(srcPath), ".tif") {
		img, err = tiff.Decode(f)
	} else {
		img, _, err = image.Decode(f)
	}
	if err != nil {
		return err
	}

	rgba := toRGBA(img)
	if removeWhite {
		floodWhite(rgba)
	}
	rgba = trim(rgba)
	rgba = scale(rgba)

	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer out.Close()
	return png.Encode(out, rgba)
}

func toRGBA(img image.Image) *image.RGBA {
	b := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), img, b.Min, draw.Src)
	return dst
}

// floodWhite makes the BACKGROUND transparent via BFS from every border pixel:
// only background-colored regions CONNECTED TO THE EDGE are cleared, so
// same-colored details inside the mascot (white eyes) are preserved. The
// background color is sampled from the corners — white on the brand TIFs,
// BLACK on the animation videos.
func floodWhite(img *image.RGBA) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	bg := cornerColor(img)
	visited := make([]bool, w*h)
	queue := make([]int, 0, w*2+h*2)

	push := func(x, y int) {
		i := y*w + x
		if visited[i] || !nearColor(img.RGBAAt(x, y), bg) {
			return
		}
		visited[i] = true
		queue = append(queue, i)
	}
	for x := 0; x < w; x++ {
		push(x, 0)
		push(x, h-1)
	}
	for y := 0; y < h; y++ {
		push(0, y)
		push(w-1, y)
	}
	for len(queue) > 0 {
		i := queue[0]
		queue = queue[1:]
		x, y := i%w, i/w
		img.SetRGBA(x, y, color.RGBA{})
		if x > 0 {
			push(x-1, y)
		}
		if x < w-1 {
			push(x+1, y)
		}
		if y > 0 {
			push(x, y-1)
		}
		if y < h-1 {
			push(x, y+1)
		}
	}
}

// cornerColor samples the four corners and returns their average — the
// background key color.
func cornerColor(img *image.RGBA) color.RGBA {
	b := img.Bounds()
	cs := []color.RGBA{
		img.RGBAAt(b.Min.X, b.Min.Y),
		img.RGBAAt(b.Max.X-1, b.Min.Y),
		img.RGBAAt(b.Min.X, b.Max.Y-1),
		img.RGBAAt(b.Max.X-1, b.Max.Y-1),
	}
	var r, g, bb int
	for _, c := range cs {
		r += int(c.R)
		g += int(c.G)
		bb += int(c.B)
	}
	return color.RGBA{uint8(r / 4), uint8(g / 4), uint8(bb / 4), 0xff}
}

// nearColor reports whether c is within tolerance of the background key.
func nearColor(c, key color.RGBA) bool {
	const tol = 34
	d := func(a, b uint8) int {
		if a > b {
			return int(a - b)
		}
		return int(b - a)
	}
	return d(c.R, key.R) <= tol && d(c.G, key.G) <= tol && d(c.B, key.B) <= tol
}

// trim crops to the alpha bounding box plus padding.
func trim(img *image.RGBA) *image.RGBA {
	b := img.Bounds()
	minX, minY, maxX, maxY := b.Max.X, b.Max.Y, b.Min.X, b.Min.Y
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.RGBAAt(x, y).A > 8 {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}
	if minX > maxX { // fully transparent — keep as is
		return img
	}
	minX = max(b.Min.X, minX-trimPad)
	minY = max(b.Min.Y, minY-trimPad)
	maxX = min(b.Max.X-1, maxX+trimPad)
	maxY = min(b.Max.Y-1, maxY+trimPad)

	r := image.Rect(0, 0, maxX-minX+1, maxY-minY+1)
	dst := image.NewRGBA(r)
	draw.Draw(dst, r, img, image.Pt(minX, minY), draw.Src)
	return dst
}

func scale(img *image.RGBA) *image.RGBA {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= maxDim && h <= maxDim {
		return img
	}
	s := float64(maxDim) / float64(w)
	if hs := float64(maxDim) / float64(h); hs < s {
		s = hs
	}
	dst := image.NewRGBA(image.Rect(0, 0, int(float64(w)*s), int(float64(h)*s)))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, b, draw.Src, nil)
	return dst
}
