#!/bin/sh
set -eu

tmpdir="$(mktemp -d "${TMPDIR:-/tmp}/schedulex-integration.XXXXXX")"
cleanup() {
  rm -rf "$tmpdir"
}
trap cleanup EXIT INT TERM

# Local downstream smoke uses the committed module identity. The fixture go.mod has no local replace;
# CI can run it after v0.1.0 publication. Local pre-release verifies the import path via examples
# plus rendered downstream modules.
! grep -q '^replace ' test/downstream-smoke/go.mod
GOWORK=off go test ./examples/...

render_and_check() {
  name="$1"
  module_path="$2"
  package_name="$3"
  out="$tmpdir/$name"

  scripts/render_template.sh \
    --module-name "$name" \
    --module-path "$module_path" \
    --package-name "$package_name" \
    --out "$out"
  scripts/check_rendered_template.sh "$out" "$name" "$module_path" "$package_name"
  (cd "$out" && GOWORK=off go test ./...)
}

render_and_check kernel github.com/ZoneCNH/kernel kernel
render_and_check corekit example.com/acme/corekit corekit
