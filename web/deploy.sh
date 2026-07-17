#!/bin/sh
# Stage the web bundle and deploy to Cloudflare Pages (free tier).
#
# One-time setup:
#   npm install -g wrangler && wrangler login
#
# (Alternative: `npx wrangler deploy` from the repo root deploys the same
# bundle as a Worker instead — see wrangler.jsonc.)
set -e
cd "$(dirname "$0")/.."
./web/stage.sh
npx wrangler pages deploy web --project-name sapootchi
