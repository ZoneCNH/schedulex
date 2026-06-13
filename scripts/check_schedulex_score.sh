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

export VERSION="${VERSION:-v1.0.0}"
export GOWORK="${GOWORK:-off}"

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
deductions="0.0"

add_deduction() {
  deductions=$(awk -v current="$deductions" -v amount="$1" 'BEGIN { printf "%.1f", current + amount }')
}

# 检查 go vet
if ! GOWORK=off go vet ./... 2>/dev/null; then
  add_deduction 2.0
fi

# 检查测试通过
if ! GOWORK=off go test ./... 2>/dev/null; then
  add_deduction 2.0
fi

# 检查覆盖率（低于 80% 扣分）
coverage=$(GOWORK=off go test ./pkg/schedulex -coverprofile=/tmp/_score_cover.out 2>/dev/null | awk '/coverage:/ {
  for (i = 1; i <= NF; i++) {
    if ($i ~ /^[0-9]+(\.[0-9]+)?%$/) {
      gsub("%", "", $i)
      print $i
      exit
    }
  }
}')
if [[ -n "$coverage" ]] && awk -v coverage="$coverage" 'BEGIN { exit (coverage + 0 < 80) ? 0 : 1 }'; then
  add_deduction 1.0
fi

# 检查 race
if ! GOWORK=off go test -race ./pkg/schedulex 2>/dev/null; then
  add_deduction 1.0
fi

score=$(awk -v deductions="$deductions" 'BEGIN { printf "%.1f", 10.0 - deductions }')

awk -v s="$score" -v m="$min" 'BEGIN { exit (s + 0 >= m + 0) ? 0 : 1 }' || {
  echo "ERROR: score=$score below min=$min" >&2
  exit 1
}

echo "score=$score min=$min status=pass"
