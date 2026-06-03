# Downstream Registry

Goal v2.9.3 的代表下游以 patch-only 方式验证，不直接写外部仓库。

| Repo | Role | Mode | Notes |
| --- | --- | --- | --- |
| `kernel/configx` | L0 representative downstream | `patch-only` | 用于 `downstream-baseline` 与 `downstream-adoption` dry-run。 |
| `kernel` | integration smoke | `patch-only` | 保留既有 integration 代表下游语境。 |

`downstream-baseline --repo kernel/configx --mode patch-only` 与 `downstream-adoption --repo kernel/configx --mode patch-only` 只产生命令报告；真实 adoption 需要独立 PR/Evidence。
