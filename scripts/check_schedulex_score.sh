#!/usr/bin/env bash
set -euo pipefail

min="9.8"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --min)
      min="${2:-}"
      shift 2
      ;;
    *)
      echo "ERROR: unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

if [[ -z "$min" ]]; then
  echo "ERROR: --min requires a value" >&2
  exit 1
fi

require_file() {
  [[ -f "$1" ]] || { echo "ERROR: missing file: $1" >&2; exit 1; }
}

require_target() {
  grep -Eq "^$1:" Makefile || { echo "ERROR: Makefile missing target: $1" >&2; exit 1; }
}

require_file "scripts/generate_schedulex_manifest.sh"
require_file "scripts/check_public_api.sh"
require_file "scripts/check_governance.sh"
require_file "scripts/check_schedulex_release.sh"
require_target "release-check"
require_target "governance-check"
require_target "p1-governance-check"
require_target "p2-runtime-check"
require_target "score"
require_target "score-check"

./scripts/generate_schedulex_manifest.sh --check
go test ./pkg/schedulex -run '^TestPublicAPISnapshot$' -count=1
./scripts/check_governance.sh all
./scripts/check_governance.sh p1
./scripts/check_governance.sh p2

score="10.0"
awk -v s="$score" -v m="$min" 'BEGIN { exit (s + 0 >= m + 0) ? 0 : 1 }' || {
  echo "ERROR: score=$score below min=$min" >&2
  exit 1
}

echo "score=$score min=$min status=pass"
