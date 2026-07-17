#!/bin/sh
# Build the web bundle and deploy it to Cloudflare Pages (free tier).
#
# One-time setup:
#   npm install -g wrangler
#   wrangler login
#
# Then: ./web/deploy.sh   (first run creates the "sapootchi" Pages project)
#
# Size note: Pages has a 25 MiB per-file limit. The WASM stays under it because
# animations are NOT embedded on web — they deploy as loose files under
# web/sprites/anims and are fetched at runtime.
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
	echo "ERROR: sapootchi.wasm exceeds Cloudflare Pages' 25 MiB file limit" >&2
	exit 1
fi

npx wrangler pages deploy web --project-name sapootchi
