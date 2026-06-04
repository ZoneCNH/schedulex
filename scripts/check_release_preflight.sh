#!/bin/sh
set -eu
version="${1:-v0.1.0}"
[ "$version" = "v0.1.0" ] || { echo "unexpected version $version"; exit 1; }
[ -s release/manifest/latest.json ] || ./scripts/generate_manifest.sh "$version"
grep -q '"module": "github.com/ZoneCNH/schedulex"' release/manifest/latest.json
grep -q '"version": "v0.1.0"' release/manifest/latest.json
sha256sum -c release/manifest/latest.json.sha256
