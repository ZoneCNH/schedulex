# Schedulex v1.0.0 Verification and Omission Scan Plan

This plan is the worker-3 verification artifact for the v1.0.0 release lane. It is intentionally non-mutating except where a command explicitly generates release evidence; do not commit, push, merge, or tag from a worker lane.

## Stop condition

The release is ready only when all checks below pass on the release branch, the generated manifest and checksum are present and verified, and the omission scans find no live v0.1.0/template-origin anchors outside documented historical or migration records.

## 1. Fresh workspace and core Go validation

Run from the repository root with `GOWORK=off` unless a command says otherwise.

```bash
git status --short
GOWORK=off go vet ./...
GOWORK=off go test ./...
GOWORK=off go test -race ./...
GOWORK=off go test ./pkg/schedulex -coverprofile=coverage.out
go tool cover -func=coverage.out
```

Expected evidence:
- `git status --short` shows only intentional release artifacts or documentation changes.
- `go vet`, `go test`, and race tests pass.
- Package coverage does not regress from the release-context baseline of 98.2% for `pkg/schedulex`.

## 2. Lint, contracts, and dedicated release gates

```bash
GOWORK=off make lint
GOWORK=off make contracts
GOWORK=off make api-check
GOWORK=off make trigger-determinism-check
GOWORK=off make misfire-contract-check
GOWORK=off make timezone-dst-golden-check
GOWORK=off make scheduler-leak-check
GOWORK=off make scheduler-race-check
GOWORK=off make lock-interface-check
```

Expected evidence:
- All make targets exit 0.
- The release environment has required tools installed before the final run: `golangci-lint`, `govulncheck`, `bc`, `sha256sum`, and Go 1.23+.

## 3. Manifest, release evidence, score, and preflight

Use `VERSION=v1.0.0` for every release command so generated metadata and downstream checks cannot silently default to an older tag.

```bash
VERSION=v1.0.0 GOWORK=off make evidence
VERSION=v1.0.0 GOWORK=off make release-final-check
VERSION=v1.0.0 GOWORK=off make score
VERSION=v1.0.0 GOWORK=off make release-preflight
VERSION=v1.0.0 GOWORK=off make release-check
```

Non-mutating readiness probes before generated files exist:

```bash
VERSION=v1.0.0 ./scripts/generate_schedulex_manifest.sh --check
./scripts/check_release_preflight.sh v1.0.0
./scripts/check_schedulex_release.sh
./scripts/check_schedulex_score.sh
./scripts/check_docs.sh
```

Expected evidence:
- `release/manifest/latest.json` and `release/manifest/latest.json.sha256` both exist, and stale template inputs such as `release/manifest/template.json` cannot overwrite the schedulex manifest shape.
- `scripts/generate_schedulex_manifest.sh --check` verifies the manifest checksum.
- Score meets the v1.0.0 full-score threshold.
- Any generated evidence files are reviewed before commit by the release owner.

## 4. Omission scans before declaring v1.0.0 ready

### 4.1 Live stale-version anchors

```bash
rg -n --hidden \
  --glob '!vendor/**' \
  --glob '!.git/**' \
  --glob '!coverage.out' \
  --glob '!CHANGELOG.md' \
  --glob '!docs/adr/**' \
  --glob '!docs/migration/**' \
  --glob '!release/evidence/goalcli/**' \
  'v0\.1\.0' \
  AGENTS.md CONSTITUTION.md README.md docs .agent Makefile scripts contracts release pkg
```

Expected evidence:
- No live release/default/config/API/manifest/downstream references remain at `v0.1.0`.
- Historical references are kept only in clearly historical files such as changelog, ADRs, migration notes, or archived evidence.

Current blocking examples found by worker-3 before this plan was added:
- `Makefile` default `VERSION ?= v0.1.0`.
- `pkg/schedulex/scheduler.go` public `Version` constant.
- `contracts/public_api.snapshot` and `contracts/release_manifest.schema.json`.
- `scripts/generate_manifest.sh`, `scripts/generate_schedulex_manifest.sh`, and `scripts/check_downstream_smoke.sh` defaults.
- `README.md`, `docs/release.md`, `.agent/harness.yaml`, `release/downstream-adoption/latest.json`, `release/downstream-adoption/fixture/go.mod`, and `test/downstream-smoke/go.mod`.

### 4.2 Template/provenance identity anchors

```bash
rg -n --hidden \
  --glob '!vendor/**' \
  --glob '!.git/**' \
  --glob '!CHANGELOG.md' \
  --glob '!docs/adr/**' \
  --glob '!docs/migration/**' \
  --glob '!release/evidence/goalcli/**' \
  'xlib-standard|templatex|baselib-template|cmd/goalcli|pkg/templatex' \
  AGENTS.md CONSTITUTION.md README.md docs .agent Makefile scripts contracts release pkg
```

Expected evidence:
- Live governance, score, and harness files describe schedulex, not `xlib-standard`, `templatex`, `baselib-template`, `cmd/goalcli`, or `pkg/templatex`.
- Any retained template-origin wording is explicitly marked as historical provenance and is not part of a required release gate.

Current blocking examples found by worker-3 before this plan was added:
- `CONSTITUTION.md` title and content still reference `xlib-standard` and `cmd/goalcli`.
- `.agent/release-required-gates.yaml` still names module `xlib-standard` and contains `cmd/goalcli` score gates.
- `docs/scorecard.md`, `docs/standard/release-standard.md`, and `docs/standard/harness-gates.md` reference non-existent command paths such as `cmd/schedulex` or legacy `cmd/goalcli` in live release-score documentation.

## 5. Post-tag adoption check

Run only after the `v1.0.0` tag is published and visible to the Go proxy or direct module fetch path.

```bash
tmpdir="$(mktemp -d)"
cd "$tmpdir"
go mod init example.com/schedulex-release-check
go get github.com/ZoneCNH/schedulex@v1.0.0
cat > main.go <<'GO'
package main

import _ "github.com/ZoneCNH/schedulex/pkg/schedulex"

func main() {}
GO
GOWORK=off go test ./...
```

Expected evidence:
- The module resolves at `github.com/ZoneCNH/schedulex@v1.0.0` without local `replace` directives.
- The smoke import compiles in a clean module.

## 6. Command-order attribution hazards

When a score or release gate fails, record the root-cause command rather than only the aggregate target. In particular, score failures can originate from manifest absence, stale version defaults, governance drift, missing tool prerequisites, or non-existent command paths referenced by docs/gates.

## 7. Release-owner handoff checklist

Before tagging, confirm:
- All version sources and generated metadata use `v1.0.0`.
- `release/manifest/latest.json` and its checksum are present, verified, and committed by the release owner.
- `release/manifest/template.json`, downstream fixtures, and smoke-test modules are either updated for v1.0.0 or explicitly excluded from live release gates as historical/reference-only artifacts.
- Release gates do not call missing command paths.
- Full-score documentation and CI/harness files agree on the same commands.
- The final release branch is clean after generated evidence review.
