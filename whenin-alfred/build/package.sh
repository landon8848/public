#!/bin/sh
# Build a universal Go binary and package the Alfred workflow.
#
# Produces WhenIn.alfredworkflow (a zip with info.plist at its root,
# alongside the compiled `whenin` binary and icon). Run from anywhere:
#   sh build/package.sh
set -eu

HERE=$(cd "$(dirname "$0")" && pwd)
ROOT=$(dirname "$HERE")
SRC="$ROOT/workflow"
BIN="$SRC/whenin"
OUT="$ROOT/WhenIn.alfredworkflow"

if [ ! -f "$ROOT/internal/whenin/index.json" ]; then
	echo "internal/whenin/index.json missing — run: python3 build/build_index.py" >&2
	exit 1
fi

# Universal binary (arm64 + amd64) so the workflow runs on any Mac.
echo "Building arm64…"
( cd "$ROOT" && GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o "$BIN.arm64" . )
echo "Building amd64…"
( cd "$ROOT" && GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o "$BIN.amd64" . )
lipo -create -output "$BIN" "$BIN.arm64" "$BIN.amd64"
rm -f "$BIN.arm64" "$BIN.amd64"
chmod +x "$BIN"

rm -f "$OUT"
# Zip the *contents* of workflow/ so info.plist sits at the archive root.
( cd "$SRC" && zip -r -q -X "$OUT" . -x '.DS_Store' )
rm -f "$BIN"

SIZE=$(du -h "$OUT" | cut -f1)
echo "Built $OUT ($SIZE)"
