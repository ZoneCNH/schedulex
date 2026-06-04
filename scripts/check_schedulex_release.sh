#!/usr/bin/env bash
set -euo pipefail

./scripts/check_public_api.sh
go test ./pkg/schedulex ./examples ./contracts
./scripts/check_downstream_smoke.sh
./scripts/generate_schedulex_manifest.sh --check
