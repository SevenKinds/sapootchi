#!/bin/sh
# Build the WASM bundle and serve it on the local network so a phone on the
# same Wi-Fi can play: run this, then open the printed URL on the phone.
set -e
cd "$(dirname "$0")/.."

echo "building wasm..."
GOOS=js GOARCH=wasm go build -o web/sapootchi.wasm .

# wasm_exec.js ships with the Go toolchain (lib/wasm since Go 1.24).
WASM_EXEC="$(go env GOROOT)/lib/wasm/wasm_exec.js"
[ -f "$WASM_EXEC" ] || WASM_EXEC="$(go env GOROOT)/misc/wasm/wasm_exec.js"
cp "$WASM_EXEC" web/wasm_exec.js

IP=$(ipconfig getifaddr en0 2>/dev/null || ipconfig getifaddr en1 2>/dev/null || echo localhost)
echo ""
echo "  On your phone (same Wi-Fi):  http://$IP:8080"
echo "  On this machine:             http://localhost:8080"
echo ""
cd web && exec python3 -m http.server 8080 --bind 0.0.0.0
