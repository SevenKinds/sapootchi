#!/bin/sh
# Stage the web bundle and serve it on the local network so a phone on the
# same Wi-Fi can play: run this, then open the printed URL on the phone.
set -e
cd "$(dirname "$0")/.."
./web/stage.sh

IP=$(ipconfig getifaddr en0 2>/dev/null || ipconfig getifaddr en1 2>/dev/null || echo localhost)
echo ""
echo "  On your phone (same Wi-Fi):  http://$IP:8080"
echo "  On this machine:             http://localhost:8080"
echo ""
cd web && exec python3 -m http.server 8080 --bind 0.0.0.0
