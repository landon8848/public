#!/bin/sh
# Package the Alfred workflow — no binary included.
#
# Produces WhenIn.alfredworkflow (a zip with info.plist at its root,
# alongside the icon). The compiled engine ships separately: users install
# it with `go install github.com/landon8848/public/whenin-alfred@latest`,
# and the workflow's script filter finds it at runtime. Building locally
# also sidesteps Gatekeeper, since a binary you compile yourself is never
# quarantined. Run from anywhere:
#   sh build/package.sh
set -eu

HERE=$(cd "$(dirname "$0")" && pwd)
ROOT=$(dirname "$HERE")
SRC="$ROOT/workflow"
OUT="$ROOT/WhenIn.alfredworkflow"

rm -f "$OUT"
# Zip the *contents* of workflow/ so info.plist sits at the archive root.
# Anything stray (a locally built binary, .DS_Store) is left out on purpose.
( cd "$SRC" && zip -r -q -X "$OUT" info.plist icon.png )

SIZE=$(du -h "$OUT" | cut -f1)
echo "Built $OUT ($SIZE)"
