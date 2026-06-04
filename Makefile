SHELL := /bin/sh
VERSION ?= v0.1.0

.PHONY: require-gowork-off fmt vet test race build boundary contracts golden property schedulex-checks release-preflight ci downstream-smoke manifest lint

require-gowork-off:
	@if [ "$${GOWORK:-}" != "off" ]; then echo "GOWORK=off is required"; exit 1; fi

fmt:
	go fmt ./...

vet:
	go vet ./...

lint: vet

test:
	go test ./...

race:
	go test -race ./pkg/schedulex

build:
	go build ./...

boundary: require-gowork-off
	./scripts/check_boundary.sh

contracts: require-gowork-off
	./scripts/check_contracts.sh

golden: require-gowork-off
	go test ./pkg/schedulex ./contracts -run 'Golden|Snapshot|Contracts'

property: require-gowork-off
	go test ./pkg/schedulex -run 'Deterministic|Contract|DST|Concurrency|Shutdown|Leak'

manifest: require-gowork-off
	./scripts/generate_manifest.sh $(VERSION)

schedulex-checks: require-gowork-off fmt vet test boundary contracts golden property race manifest

release-preflight: require-gowork-off schedulex-checks
	./scripts/check_release_preflight.sh $(VERSION)

ci: require-gowork-off schedulex-checks

downstream-smoke: require-gowork-off
	./scripts/run_integration.sh
