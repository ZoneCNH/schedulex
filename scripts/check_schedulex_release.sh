#!/usr/bin/env bash
set -euo pipefail

version="${VERSION:-v1.0.0}"
export VERSION="$version"
export GOWORK="${GOWORK:-off}"

./scripts/check_public_api.sh
go test ./pkg/schedulex ./examples ./contracts
./scripts/check_downstream_smoke.sh
./scripts/generate_schedulex_manifest.sh --check
