//go:build js

package game

import (
	"bytes"
	"fmt"
	"image"
	"syscall/js"

	"github.com/hajimehoshi/ebiten/v2"

	"sapootchi/assets"
)

// loadAnims (web): the animations are NOT embedded (they would push the WASM
// past Cloudflare Pages' 25 MiB file limit). Fetch them in the background via
// the browser's native fetch (net/http costs ~8 MiB of WASM; syscall/js is
// free) — each finished animation lands in the game loop via the channel and
// becomes tappable the moment it arrives.
func loadAnims(b *spriteBank) {
	b.animsIn = make(chan animDelivery, len(assets.AnimNames))
	go func() {
		for _, name := range assets.AnimNames {
			var seq []*ebiten.Image
			for i := 0; ; i++ {
				data, ok := fetchBytes(fmt.Sprintf("sprites/anims/%s/%03d.png", name, i))
				if !ok {
					break
				}
				img, _, err := image.Decode(bytes.NewReader(data))
				if err != nil {
					break
				}
				seq = append(seq, ebiten.NewImageFromImage(img))
			}
			if len(seq) > 0 {
				b.animsIn <- animDelivery{name: name, frames: seq}
			}
		}
	}()
}

// fetchBytes GETs a same-origin URL with the browser fetch API. Blocks the
// calling goroutine (fine — it runs off the game loop).
func fetchBytes(url string) (data []byte, ok bool) {
	done := make(chan struct{})

	var onBuf, onResp, onErr js.Func
	release := func() {
		onBuf.Release()
		onResp.Release()
		onErr.Release()
	}

	onBuf = js.FuncOf(func(_ js.Value, args []js.Value) any {
		u8 := js.Global().Get("Uint8Array").New(args[0])
		data = make([]byte, u8.Get("length").Int())
		js.CopyBytesToGo(data, u8)
		ok = true
		close(done)
		return nil
	})
	onResp = js.FuncOf(func(_ js.Value, args []js.Value) any {
		resp := args[0]
		if !resp.Get("ok").Bool() {
			close(done)
			return nil
		}
		resp.Call("arrayBuffer").Call("then", onBuf)
		return nil
	})
	onErr = js.FuncOf(func(_ js.Value, _ []js.Value) any {
		close(done)
		return nil
	})

	js.Global().Call("fetch", url).Call("then", onResp).Call("catch", onErr)
	<-done
	release()
	return data, ok
}
