#!/bin/sh
set -eu
# Local downstream smoke uses the committed module identity. The fixture go.mod has no local replace;
# CI can run it after v0.1.0 publication. Local pre-release verifies the import path via examples.
! grep -q '^replace ' test/downstream-smoke/go.mod
go test ./examples/...
