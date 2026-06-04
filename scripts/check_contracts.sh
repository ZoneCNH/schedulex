#!/bin/sh
set -eu
for f in contracts/public_api.snapshot contracts/trigger_cases/l1_golden.json contracts/trigger_cases/dst_golden.json contracts/misfire_cases/l1_golden.json contracts/lifecycle_event.schema.json contracts/release_manifest.schema.json; do
  [ -s "$f" ] || { echo "missing contract $f"; exit 1; }
done
go test ./contracts ./pkg/schedulex -run 'Contract|Golden|Snapshot|DST'
