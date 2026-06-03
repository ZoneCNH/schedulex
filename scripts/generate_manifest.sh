#!/usr/bin/env bash
set -euo pipefail

go run ./internal/tools/releasemanifest --out release/manifest/latest.json
