# Gate

## Required Gates

- Format Gate：`GOWORK=off make fmt`
- Static Check Gate：`GOWORK=off make vet`
- Lint Gate：`GOWORK=off make lint`，缺少 `golangci-lint` 时失败
- Unit Test Gate：`GOWORK=off make test`
- Race Test Gate：`GOWORK=off make race`
- Boundary Gate：`GOWORK=off make boundary`
- Secret Gate：`GOWORK=off make security`，必须包含 `govulncheck ./...` 和密钥扫描
- Contract Gate：`GOWORK=off make contracts`
- Docs Gate：`GOWORK=off make docs-check`
- Integration Gate：`GOWORK=off make integration`，默认下游为 `kernel`
- Evidence Gate：`CHECK_STATUS=passed GOWORK=off make evidence`
- Release Gate：`GOWORK=off make release-check`

## Final Gates

- `XLIB_CONTEXT=release_verify GOWORK=off make release-final-check`
- `XLIB_CONTEXT=release_verify GOWORK=off make release-preflight VERSION=<version>`
- `goalcli score --min 9.8`
- kernel downstream smoke：渲染后执行 `GOWORK=off go test ./...`、`make contracts`、`make boundary` 和 release Evidence 校验。

## Extended Gates

- Property Gate：`GOWORK=off make property`
- Fuzz Smoke Gate：`FUZZ_SMOKE_TIME=<duration> GOWORK=off make fuzz-smoke`
- Golden Gate：`GOWORK=off make golden`
- Extended CI Gate：`GOWORK=off make ci-extended`
- Extended Release Gate：`GOWORK=off make release-check-extended`

## Policy

Required Gates 是 `xlib-standard` 和所有生成基础库的强制基线。Extended Gates 推荐所有生成基础库实现，并对 storage、messaging、observability 和 security-sensitive 基础库强制执行。Chaos、mutation 和 long soak 等 profile-specific heavy gates 不进入默认 `make ci`。


## Goal v2.9.3 Governance Gates

- P0 Governance Gate：`XLIB_CONTEXT=local_write GOWORK=off make governance-check`，串联 `main-guard`、`worktree-guard`、`evidence-check`、`boundary`、`security`、`cli-contract`、`issue-registry`、`command-registry` 和 `makefile-baseline`。
- P1 Governance Dry Run：`GOWORK=off make p1-governance-check`，只做本地文档/registry/schema 证明，不读取 GitHub secrets，不写外部路径。
- P2 Runtime Dry Run：`GOWORK=off make p2-runtime-check`，验证 conformance profile、pack contract、downstream patch-only、runtime-file-ownership 和 execution-context 文档契约。
