#!/usr/bin/env bash
# lem-in visualizer launcher.
#
# Pipe lem-in's output straight into it:
#     ./lem-in tests/example00.txt | visualizer/serve.sh
#
# ...or let it run lem-in for you:
#     visualizer/serve.sh tests/example00.txt
#
# Then open http://localhost:8000
set -e
cd "$(dirname "$0")"

if [ $# -ge 1 ]; then
    ../lem-in "$1" > output.txt
else
    cat > output.txt
fi

# sanity check: the file must contain at least one turn
if ! grep -q '^L[0-9]' output.txt; then
    echo "ERROR: no moves found in input — is this lem-in output?" >&2
    exit 1
fi

PORT="${PORT:-8000}"
echo "🐜 visualizer serving at http://localhost:${PORT}  (Ctrl+C to stop)"
( sleep 1; xdg-open "http://localhost:${PORT}" >/dev/null 2>&1 || true ) &
python3 -m http.server "$PORT" >/dev/null 2>&1
