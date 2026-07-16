//go:build dev

package game

// DevMode is on in dev builds:
//
//	go run -tags dev .
//	GOOS=js GOARCH=wasm go build -tags dev -o web/sapootchi.wasm .
//
// It adds a DEV section to Settings (spawn Energy Pills, grant coins). Dev
// items live in the normal save, so switching back to a release build keeps
// them usable but unobtainable.
const DevMode = true
