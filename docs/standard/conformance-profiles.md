# Conformance Profiles

| Profile | Scope | Command |
| --- | --- | --- |
| `standard-source` | xlib-standard 作为标准源、模板、generator、Harness 和 Evidence 实现仓库 | `goalcli attest-conformance --profile standard-source` |
| `l0-kernel` | L0 representative downstream 与 kernel/configx patch-only 验证 | `goalcli attest-conformance --profile l0-kernel` |

Profile 验证是 dry-run contract check，不读取真实 secrets，不写下游仓库。
