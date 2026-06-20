#!/bin/sh
# Render the workflow icon from the source SVG.
# Requires librsvg (`brew install librsvg`).  Run from anywhere:
#   sh build/make_icon.sh
set -eu

HERE=$(cd "$(dirname "$0")" && pwd)
ROOT=$(dirname "$HERE")
SRC="$HERE/whenin_icon.svg"
OUT="$ROOT/workflow/icon.png"

if ! command -v rsvg-convert >/dev/null 2>&1; then
	echo "rsvg-convert not found — install with: brew install librsvg" >&2
	exit 1
fi

rsvg-convert -w 512 -h 512 "$SRC" -o "$OUT"
echo "Wrote $OUT ($(du -h "$OUT" | cut -f1))"
