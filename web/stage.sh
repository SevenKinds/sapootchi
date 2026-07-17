#!/bin/sh
# Build the stripped WASM and stage everything the web bundle needs.
set -e
cd "$(dirname "$0")/.."

echo "building wasm (stripped)..."
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o web/sapootchi.wasm .

WASM_EXEC="$(go env GOROOT)/lib/wasm/wasm_exec.js"
[ -f "$WASM_EXEC" ] || WASM_EXEC="$(go env GOROOT)/misc/wasm/wasm_exec.js"
cp "$WASM_EXEC" web/wasm_exec.js

echo "staging runtime-fetched animations..."
rm -rf web/sprites
mkdir -p web/sprites
cp -R assets/sprites/anims web/sprites/anims

SIZE=$(wc -c < web/sapootchi.wasm)
echo "wasm size: $((SIZE / 1048576)) MiB (limit 25)"
if [ "$SIZE" -gt 26214400 ]; then
	echo "ERROR: sapootchi.wasm exceeds Cloudflare's 25 MiB file limit" >&2
	exit 1
fi
