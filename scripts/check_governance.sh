#!/usr/bin/env bash
set -euo pipefail

mode="${1:-all}"

fail() {
  echo "ERROR: $*" >&2
  exit 1
}

require_file() {
  [[ -f "$1" ]] || fail "missing file: $1"
}

require_text() {
  grep -Fq -- "$2" "$1" || fail "$1 missing text: $2"
}

require_make_target() {
  grep -Eq "^$1:" Makefile || fail "Makefile missing target: $1"
}

check_all() {
  require_file ".agent/goal-runtime.md"
  require_file ".agent/harness.yaml"
  require_file ".agent/traceability-matrix.md"
  require_file ".agent/release-required-gates.yaml"
  require_file ".agent/command-registry.yaml"
  require_file ".agent/makefile-target-registry.yaml"
  require_file ".agent/acceptance-matrix.yaml"
  require_file ".agent/runtime-health.yaml"
  require_file "docs/standard/harness-gates.md"
  require_file "docs/standard/schedulex.md"
  require_file "contracts/release_manifest.schema.json"
  require_file "release/manifest/latest.json"
  require_file "release/manifest/latest.json.sha256"
  require_file ".github/workflows/ci.yml"
  require_file ".github/pull_request_template.md"

  require_make_target "governance-check"
  require_make_target "p1-governance-check"
  require_make_target "p2-runtime-check"
  require_make_target "score"
  require_make_target "score-check"
  require_make_target "release-check"
  require_make_target "ci-extended"
  require_make_target "release-preflight"

  require_text ".agent/release-required-gates.yaml" "GOWORK=off make governance-check"
  require_text ".agent/release-required-gates.yaml" "GOWORK=off make p1-governance-check"
  require_text ".agent/release-required-gates.yaml" "GOWORK=off make p2-runtime-check"
  require_text ".agent/release-required-gates.yaml" "score"
  require_text ".agent/release-required-gates.yaml" "GOWORK=off make release-check"
  require_text ".github/workflows/ci.yml" "make release-check"
  require_text "docs/standard/harness-gates.md" "GOWORK=off make governance-check"
  require_text "docs/standard/harness-gates.md" "GOWORK=off make p1-governance-check"
  require_text "docs/standard/harness-gates.md" "GOWORK=off make p2-runtime-check"
  require_text "docs/standard/harness-gates.md" "GOWORK=off make score"

  (cd release/manifest && sha256sum -c latest.json.sha256 >/dev/null)
}

check_p1() {
  require_file ".agent/acceptance-matrix.yaml"
  require_file ".agent/command-implementation-status.yaml"
  require_file ".agent/makefile-baseline.yaml"
  require_file ".agent/pr-template-contract.yaml"
  require_file ".github/pull_request_template.md"
  require_file ".github/workflows/ci.yml"
  require_file ".agent/toolchain.yaml"

  require_text ".agent/toolchain.yaml" "golangci-lint"
  require_text ".agent/toolchain.yaml" "govulncheck"
  require_text ".github/workflows/ci.yml" "golangci-lint"
  require_text ".github/workflows/ci.yml" "govulncheck"
}

check_p2() {
  require_file ".agent/runtime-health.yaml"
  require_file ".agent/install-runtime.md"
  require_file ".agent/upgrade-runtime.md"
  require_file ".agent/execution-context.yaml"
  require_file ".agent/downstream-adoption-status.yaml"
  require_file "release/downstream-adoption/fixture/go.mod"
  require_file "scripts/check_downstream_smoke.sh"

  if grep -Eq '^[[:space:]]*replace[[:space:]]' release/downstream-adoption/fixture/go.mod; then
    fail "downstream fixture must not use a local replace"
  fi
}

case "$mode" in
  all)
    check_all
    ;;
  p1)
    check_p1
    ;;
  p2)
    check_p2
    ;;
  *)
    fail "unknown governance mode: $mode"
    ;;
esac

echo "governance check passed: $mode"
