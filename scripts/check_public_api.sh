#!/usr/bin/env bash
set -euo pipefail

go test ./pkg/schedulex -run '^TestPublicAPISnapshot$' -count=1
go test ./contracts -run '^TestPublicAPISnapshotDocumentsExportedSurface$' -count=1
