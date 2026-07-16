//go:build js

package game

import (
	"errors"
	"syscall/js"
)

// Browser (WASM) storage: the save lives in localStorage, so the pet persists
// across page reloads on the same device/browser.

const saveKey = "sapootchi_save"

var errNoSave = errors.New("no save in localStorage")

func readSave() ([]byte, error) {
	v := js.Global().Get("localStorage").Call("getItem", saveKey)
	if v.IsNull() || v.IsUndefined() {
		return nil, errNoSave
	}
	return []byte(v.String()), nil
}

func writeSave(data []byte) error {
	js.Global().Get("localStorage").Call("setItem", saveKey, string(data))
	return nil
}
