#!/bin/sh
set -eu

version="${1:-${VERSION:-}}"
if [ -z "$version" ]; then
  echo "ERROR: release preflight requires VERSION=<semver> or an explicit version argument" >&2
  exit 1
fi
case "$version" in
  v[0-9]*.[0-9]*.[0-9]*) ;;
  *) echo "ERROR: unexpected release version: $version" >&2; exit 1 ;;
esac

VERSION="$version" ./scripts/generate_schedulex_manifest.sh
VERSION="$version" ./scripts/generate_schedulex_manifest.sh --check

grep -q '"module": "github.com/ZoneCNH/schedulex"' release/manifest/latest.json
grep -q "\"version\": \"$version\"" release/manifest/latest.json
(cd release/manifest && sha256sum -c latest.json.sha256)
