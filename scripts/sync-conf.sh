#!/bin/bash
set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SRC="$ROOT"
DST="$ROOT/bin"
mkdir -p "$DST"

for f in sr-c1.json repo.json; do
    if [ ! -f "$DST/$f" ] || [ "$SRC/$f" -nt "$DST/$f" ]; then
        cp -v "$SRC/$f" "$DST/$f"
    fi
done