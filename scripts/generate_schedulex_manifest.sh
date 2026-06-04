#!/usr/bin/env bash
set -euo pipefail

mode="write"
if [[ "${1:-}" == "--check" ]]; then
  mode="check"
fi

manifest="release/manifest/latest.json"
checksum="release/manifest/latest.json.sha256"
module="$(go list -m)"
version="${VERSION:-v0.1.0}"
goal_id="${GOAL_ID:-GOAL-20260604-SCHEDULEX-001}"
go_version="$(go version | tr -d '\n')"
git_commit="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"
git_tree="$(git rev-parse HEAD^{tree} 2>/dev/null || echo unknown)"
workspace_status="clean"
if ! git diff --quiet || ! git diff --cached --quiet || [[ -n "$(git ls-files --others --exclude-standard)" ]]; then
  workspace_status="dirty"
fi
generated_at="${SOURCE_DATE_EPOCH:-}"
if [[ -n "$generated_at" ]]; then
  generated_at="$(date -u -d "@$generated_at" +%Y-%m-%dT%H:%M:%SZ)"
else
  generated_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
fi

hash_file() {
  sha256sum "$1" | awk '{print $1}'
}

api_hash="$(hash_file contracts/public_api.snapshot)"
trigger_hash="$(hash_file contracts/trigger_cases/l1_golden.json)"
trigger_basic_hash="$(hash_file contracts/trigger_cases/basic.golden.json)"
dst_hash="$(hash_file contracts/trigger_cases/dst_golden.json)"
dst_basic_hash="$(hash_file contracts/timezone_dst_cases/daily_at_dst.golden.json)"
misfire_hash="$(hash_file contracts/misfire_cases/l1_golden.json)"
misfire_basic_hash="$(hash_file contracts/misfire_cases/basic.golden.json)"
lifecycle_hash="$(hash_file contracts/lifecycle_event.schema.json)"
manifest_schema_hash="$(hash_file contracts/release_manifest.schema.json)"
snapshot_schema_hash="$(hash_file contracts/snapshot.schema.json)"
downstream_hash="$(hash_file release/downstream-adoption/fixture/go.mod)"

tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT
cat > "$tmp" <<EOF_JSON
{
  "module": "$module",
  "version": "$version",
  "layer": "L1",
  "role": "deterministic_scheduler",
  "goal_id": "$goal_id",
  "standard_source": "github.com/ZoneCNH/schedulex",
  "generated_at": "$generated_at",
  "source": {
    "commit": "$git_commit",
    "tree": "$git_tree",
    "workspace_status": "$workspace_status"
  },
  "runtime": {
    "go_version": "$go_version"
  },
  "dependencies": {
    "allowed": ["go standard library"],
    "forbidden": ["x.go", "redisx", "postgresx", "kafkax", "taosx", "ossx", "clickhousex", "runtime lock implementations"]
  },
  "contracts": {
    "public_api": "contracts/public_api.snapshot",
    "trigger_golden": "contracts/trigger_cases/l1_golden.json",
    "timezone_dst_golden": "contracts/trigger_cases/dst_golden.json",
    "misfire_golden": "contracts/misfire_cases/l1_golden.json",
    "lifecycle_event_schema": "contracts/lifecycle_event.schema.json",
    "release_manifest_schema": "contracts/release_manifest.schema.json",
    "snapshot_schema": "contracts/snapshot.schema.json"
  },
  "gates": [
    "identity-check",
    "ci",
    "ci-extended",
    "fmt",
    "vet",
    "lint",
    "test",
    "race",
    "boundary",
    "contracts",
    "docs-check",
    "security",
    "evidence",
    "governance-check",
    "p1-governance-check",
    "p2-runtime-check",
    "score",
    "score-check",
    "release-check",
    "release-final-check",
    "trigger-determinism-check",
    "misfire-contract-check",
    "timezone-dst-golden-check",
    "scheduler-leak-check",
    "scheduler-race-check",
    "lock-interface-check",
    "api-check",
    "downstream-smoke"
  ],
  "downstream": {
    "fixture": "release/downstream-adoption/fixture",
    "no_local_replace": true,
    "network_publish_required_for_full_module_fetch": true
  },
  "artifacts": [
    {"path": "contracts/public_api.snapshot", "sha256": "$api_hash"},
    {"path": "contracts/trigger_cases/l1_golden.json", "sha256": "$trigger_hash"},
    {"path": "contracts/trigger_cases/basic.golden.json", "sha256": "$trigger_basic_hash"},
    {"path": "contracts/trigger_cases/dst_golden.json", "sha256": "$dst_hash"},
    {"path": "contracts/timezone_dst_cases/daily_at_dst.golden.json", "sha256": "$dst_basic_hash"},
    {"path": "contracts/misfire_cases/l1_golden.json", "sha256": "$misfire_hash"},
    {"path": "contracts/misfire_cases/basic.golden.json", "sha256": "$misfire_basic_hash"},
    {"path": "contracts/lifecycle_event.schema.json", "sha256": "$lifecycle_hash"},
    {"path": "contracts/release_manifest.schema.json", "sha256": "$manifest_schema_hash"},
    {"path": "contracts/snapshot.schema.json", "sha256": "$snapshot_schema_hash"},
    {"path": "release/downstream-adoption/fixture/go.mod", "sha256": "$downstream_hash"}
  ],
  "verification": {
    "public_api_snapshot": "GOWORK=off make api-check",
    "trigger_determinism": "GOWORK=off make trigger-determinism-check",
    "misfire_contract": "GOWORK=off make misfire-contract-check",
    "timezone_dst_golden": "GOWORK=off make timezone-dst-golden-check",
    "scheduler_leak": "GOWORK=off make scheduler-leak-check",
    "scheduler_race": "GOWORK=off make scheduler-race-check",
    "lock_interface": "GOWORK=off make lock-interface-check",
    "downstream_smoke": "GOWORK=off make downstream-smoke",
    "governance": "GOWORK=off make governance-check",
    "p1_governance": "GOWORK=off make p1-governance-check",
    "p2_runtime": "GOWORK=off make p2-runtime-check",
    "score": "GOWORK=off make score",
    "release_check": "GOWORK=off make release-check",
    "release_final": "GOWORK=off make release-final-check"
  },
  "known_gaps": [
    "full downstream module fetch requires published github.com/ZoneCNH/schedulex v0.1.0"
  ]
}
EOF_JSON

if [[ "$mode" == "check" ]]; then
  [[ -f "$manifest" ]] || { echo "missing $manifest"; exit 1; }
  [[ -f "$checksum" ]] || { echo "missing $checksum"; exit 1; }
  expected="$(sha256sum "$manifest" | awk '{print $1}')  latest.json"
  actual="$(cat "$checksum")"
  if [[ "$expected" != "$actual" ]]; then
    echo "ERROR: checksum drift: expected '$expected' got '$actual'"
    exit 1
  fi
else
  mv "$tmp" "$manifest"
  (cd "$(dirname "$manifest")" && sha256sum "$(basename "$manifest")") > "$checksum"
fi
