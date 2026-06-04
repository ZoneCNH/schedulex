#!/bin/sh
set -eu
version="${1:-v0.1.0}"
mkdir -p release/manifest
cat > release/manifest/latest.json <<JSON
{
  "module": "github.com/ZoneCNH/schedulex",
  "version": "$version",
  "artifacts": [
    "contracts/public_api.snapshot",
    "contracts/trigger_cases/l1_golden.json",
    "contracts/misfire_cases/l1_golden.json"
  ],
  "checks": [
    "GOWORK=off go test ./...",
    "GOWORK=off make boundary",
    "GOWORK=off make schedulex-checks"
  ]
}
JSON
sha256sum release/manifest/latest.json > release/manifest/latest.json.sha256
