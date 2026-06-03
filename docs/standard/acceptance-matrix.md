# Acceptance Matrix

| Tier | Gate | Command |
| --- | --- | --- |
| P0 | Governance | `XLIB_CONTEXT=local_write GOWORK=off make governance-check` |
| P0 | CLI Contract | `GOWORK=off go run ./cmd/goalcli cli-contract` |
| P0 | Registries | `GOWORK=off go run ./cmd/goalcli issue-registry && GOWORK=off go run ./cmd/goalcli command-registry` |
| P0 | Makefile Baseline | `GOWORK=off go run ./cmd/goalcli makefile-baseline` |
| P1 | Governance Dry-run | `GOWORK=off make p1-governance-check` |
| P1 | Policy Schema | `GOWORK=off go run ./cmd/goalcli policy-schema` |
| P2 | Runtime/Downstream Dry-run | `GOWORK=off make p2-runtime-check` |
| P2 | Downstream Patch-only | `GOWORK=off go run ./cmd/goalcli downstream-adoption --repo kernel/configx --mode patch-only` |

P1/P2 默认是 dry-run/patch-only，不写外部系统；真实下游采纳必须另行提供 Evidence。
