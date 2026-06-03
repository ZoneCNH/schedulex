#!/usr/bin/env bash
set -euo pipefail

args=(--verify release/manifest/latest.json)

if [[ -n "${VERSION:-}" ]]; then
  args+=(--expect-version "$VERSION")
fi

if [[ "${RELEASE_EVIDENCE_REQUIRE_PASSED:-0}" == "1" ]]; then
  args+=(--require-passed)
fi

if [[ "${RELEASE_EVIDENCE_REQUIRE_CLEAN:-0}" == "1" ]]; then
  args+=(--require-clean)
fi

if [[ -n "${RELEASE_EVIDENCE_MIN_SCORE:-}" ]]; then
  args+=(--min-score "$RELEASE_EVIDENCE_MIN_SCORE")
fi

go run ./internal/tools/releasemanifest "${args[@]}"
