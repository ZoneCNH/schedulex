#!/usr/bin/env bash
set -euo pipefail

fixture="release/downstream-adoption/fixture"
version="${VERSION:-v0.1.0}"
if grep -R --line-number '^replace ' "$fixture/go.mod"; then
  echo "ERROR: downstream smoke fixture must not use local replace"
  exit 1
fi
if [[ "${SCHEDULEX_DOWNSTREAM_NETWORK:-0}" == "1" ]]; then
  tmp="$(mktemp -d "${TMPDIR:-/tmp}/schedulex-downstream-smoke.XXXXXX")"
  cp -R "$fixture"/. "$tmp"/
  trap 'rm -rf "$tmp"' EXIT

  if ! output="$(cd "$tmp" && { GOWORK=off go mod download && GOWORK=off go test ./...; } 2>&1)"; then
    printf '%s\n' "$output"
    if printf '%s\n' "$output" | grep -q 'unknown revision'; then
      echo "ERROR: downstream smoke network mode requires the fixture version to be published first"
    fi
    exit 1
  fi
  printf '%s\n' "$output"
else
  echo "downstream smoke: PASS no local replace; network module resolution skipped (set SCHEDULEX_DOWNSTREAM_NETWORK=1 after ${version} is published)"
fi
