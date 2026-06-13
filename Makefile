SHELL := /bin/sh

VERSION ?= v1.0.0
GOAL_ID ?= GOAL-20260604-SCHEDULEX-001

.PHONY: require-gowork-off
require-gowork-off:
	@if [ "$(GOWORK)" != "off" ]; then echo "GOWORK=off is required"; exit 1; fi

.PHONY: fmt vet lint test race build
fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "missing golangci-lint"; exit 1; }
	golangci-lint run ./...

test:
	go test ./...

race:
	go test -race ./...

build:
	go build ./...

.PHONY: identity-check boundary contracts docs-check security evidence release-final-check governance-check p1-governance-check p2-runtime-check score score-check release-check
identity-check: require-gowork-off
	./scripts/check_boundary.sh

boundary: require-gowork-off
	./scripts/check_boundary.sh

contracts: require-gowork-off
	./scripts/check_contracts.sh

docs-check: require-gowork-off
	./scripts/check_docs.sh

security: require-gowork-off
	@command -v govulncheck >/dev/null 2>&1 || { echo "missing govulncheck"; exit 1; }
	govulncheck ./...
	./scripts/check_secrets.sh

evidence: require-gowork-off
	./scripts/generate_schedulex_manifest.sh

release-final-check: require-gowork-off evidence
	./scripts/check_schedulex_release.sh

governance-check: require-gowork-off evidence docs-check contracts
	./scripts/check_governance.sh all

p1-governance-check: require-gowork-off
	./scripts/check_governance.sh p1

p2-runtime-check: require-gowork-off
	./scripts/check_governance.sh p2

score: require-gowork-off evidence
	./scripts/check_schedulex_score.sh --min 9.8

score-check: score

.PHONY: trigger-determinism-check misfire-contract-check timezone-dst-golden-check
trigger-determinism-check: require-gowork-off
	go test ./pkg/schedulex ./contracts -run 'Trigger|Cron|Daily|Golden'

misfire-contract-check: require-gowork-off
	go test ./pkg/schedulex ./contracts -run 'Misfire|Contract'

timezone-dst-golden-check: require-gowork-off
	go test ./pkg/schedulex ./contracts -run 'DST|Timezone|Golden'

.PHONY: scheduler-leak-check scheduler-race-check lock-interface-check api-check downstream-smoke
scheduler-leak-check: require-gowork-off
	go test ./pkg/schedulex -run 'Leak|Shutdown'

scheduler-race-check: require-gowork-off
	go test -race ./pkg/schedulex

lock-interface-check: require-gowork-off
	go test ./pkg/schedulex ./examples/lock_interface -run 'Lock|Example|Smoke|Compiles'

api-check: require-gowork-off
	./scripts/check_public_api.sh

downstream-smoke: require-gowork-off
	./scripts/check_downstream_smoke.sh

.PHONY: integration schedulex-manifest schedulex-check schedulex-checks release-preflight ci ci-extended
integration: require-gowork-off
	./scripts/run_integration.sh

schedulex-manifest: require-gowork-off
	./scripts/generate_schedulex_manifest.sh

schedulex-check: identity-check contracts api-check trigger-determinism-check misfire-contract-check timezone-dst-golden-check scheduler-leak-check scheduler-race-check lock-interface-check downstream-smoke

schedulex-checks: schedulex-check

release-preflight: schedulex-check evidence release-final-check
	./scripts/check_release_preflight.sh "$(VERSION)"

ci: require-gowork-off fmt vet lint test race boundary contracts docs-check security

ci-extended: ci schedulex-check evidence governance-check p1-governance-check p2-runtime-check release-final-check score

release-check: ci-extended release-preflight
