#!/usr/bin/env bash
set -euo pipefail

required_files=(
  "README.md"
  "docs/standard/README.md"
  "docs/standard/schedulex.md"
  "docs/release.md"
  "docs/retrospective/schedulex-v0.1.0.md"
  "docs/downstream-sync-policy.md"
  "contracts/public_api.snapshot"
  "contracts/release_manifest.schema.json"
  "contracts/snapshot.schema.json"
  "Makefile"
  "release/manifest/latest.json"
  "release/manifest/latest.json.sha256"
)

for file in "${required_files[@]}"; do
  if [[ ! -f "$file" ]]; then
    echo "ERROR: required documentation or contract file missing: $file" >&2
    exit 1
  fi
done

require_text() {
  local file="$1"
  local needle="$2"

  if ! grep -Fq -- "$needle" "$file"; then
    echo "ERROR: $file must mention: $needle" >&2
    exit 1
  fi
}

require_text "README.md" "github.com/ZoneCNH/schedulex"
require_text "README.md" "pkg/schedulex"
require_text "README.md" "L1 deterministic scheduler"
require_text "README.md" "Clock"
require_text "README.md" "Locker"
require_text "README.md" "MisfirePolicy"
require_text "README.md" "OverlapPolicy"
require_text "README.md" "GOWORK=off make docs-check"
require_text "README.md" "GOWORK=off make release-final-check"
require_text "README.md" "release/manifest/latest.json"
require_text "README.md" "release/manifest/latest.json.sha256"
require_text "README.md" "DONE with evidence:"

require_text "docs/standard/README.md" "github.com/ZoneCNH/schedulex"
require_text "docs/standard/README.md" "docs/standard/schedulex.md"
require_text "docs/standard/README.md" "GOWORK=off make docs-check"
require_text "docs/standard/README.md" "GOWORK=off make release-preflight VERSION=v0.1.0"
require_text "docs/standard/README.md" "release/manifest/latest.json.sha256"

require_text "docs/standard/schedulex.md" "github.com/ZoneCNH/schedulex"
require_text "docs/standard/schedulex.md" "pkg/schedulex"
require_text "docs/standard/schedulex.md" "NewScheduler"
require_text "docs/standard/schedulex.md" "AddJob"
require_text "docs/standard/schedulex.md" "Start"
require_text "docs/standard/schedulex.md" "Shutdown"
require_text "docs/standard/schedulex.md" "Snapshot"
require_text "docs/standard/schedulex.md" "Once"
require_text "docs/standard/schedulex.md" "Every"
require_text "docs/standard/schedulex.md" "Cron"
require_text "docs/standard/schedulex.md" "DailyAt"
require_text "docs/standard/schedulex.md" "MisfirePolicy"
require_text "docs/standard/schedulex.md" "OverlapPolicy"
require_text "docs/standard/schedulex.md" "Clock"
require_text "docs/standard/schedulex.md" "Locker"
require_text "docs/standard/schedulex.md" "EventSink"
require_text "docs/standard/schedulex.md" "GOWORK=off make release-final-check"
require_text "docs/standard/schedulex.md" "downstream-smoke"

require_text "docs/release.md" "GOWORK=off make release-preflight VERSION=v0.1.0"
require_text "docs/release.md" "release/manifest/latest.json"
require_text "docs/release.md" "release/manifest/latest.json.sha256"

require_text "docs/retrospective/schedulex-v0.1.0.md" "schedulex gate"
require_text "docs/retrospective/schedulex-v0.1.0.md" "Locker"

require_text "docs/downstream-sync-policy.md" "schedulex"
require_text "docs/downstream-sync-policy.md" "L1 基础库"
require_text "docs/downstream-sync-policy.md" "x.go 仅作为基础库消费方"

require_text "contracts/public_api.snapshot" "func NewScheduler"
require_text "contracts/public_api.snapshot" "type Locker"
require_text "contracts/public_api.snapshot" "type Trigger"
require_text "contracts/release_manifest.schema.json" "deterministic_scheduler"
require_text "contracts/snapshot.schema.json" "jobs"

require_text "Makefile" "docs-check"
require_text "Makefile" "release-preflight"
require_text "Makefile" "downstream-smoke"

if ! (cd release/manifest && sha256sum -c latest.json.sha256 >/dev/null); then
  echo "ERROR: release manifest checksum is invalid" >&2
  exit 1
fi

echo "docs check passed"
