#!/usr/bin/env bash
set -euo pipefail

echo "checking contracts..."

REQUIRED_FILES=(
  "contracts/config.schema.json"
  "contracts/health.schema.json"
  "contracts/error.schema.json"
  "contracts/metrics.md"
  "contracts/goalcli-report.schema.json"
  "contracts/issue-registry.schema.json"
  "contracts/command-registry.schema.json"
  "contracts/execution-context.schema.json"
  "contracts/conformance-attestation.schema.json"
  "contracts/policy.schema.json"
)

for file in "${REQUIRED_FILES[@]}"; do
  if [[ ! -f "$file" ]]; then
    echo "ERROR: missing contract file: $file"
    exit 1
  fi
done

go test ./contracts

echo "contract check passed"
