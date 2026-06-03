# goalcli 单一命名合约

`goalcli` 是 Goal Runtime 的唯一命名、机器执行面和本仓库 Go 实现入口。历史旧名不再作为文档、代码、schema、Evidence 或 Makefile 的权威名称。

---

## 命名规则

| Surface | Required name |
|---|---|
| CLI 入口 | 源码入口 `cmd/goalcli`，命令名 `goalcli` |
| Makefile 变量 | `GOALCLI` |
| Report schema | `contracts/goalcli-report.schema.json` |
| Runtime 文档 | `docs/standard/goalcli-runtime.md` |
| CLI 契约文档 | `docs/standard/goalcli-cli-contract.md` |
| Evidence pack | `release/evidence/goalcli/` |
| Goal fixture ID | `GOAL-20260603-XLIB-GOALCLI-001` |

---

## 命令映射表

| Goal Runtime command | goalcli command | 状态 |
|---|---|---|
| `goalcli doctor` | `goalcli doctor` | ✅ |
| `goalcli worktree check` | `goalcli worktree-guard` | ✅ |
| `goalcli secret check` | `goalcli secrets` / `goalcli security` | ✅ |
| `goalcli schema check` | （未独立实现，融入各 check） | ⚠️ |
| `goalcli evidence check` | `goalcli evidence-check` | ✅ |
| `goalcli traceability check` | `goalcli goal-acceptance` 等 | ⚠️ |
| `goalcli release check` | `goalcli release-evidence-check` | ✅ |
| `goalcli retro check` | （隐含于 `goalcli score` 计分中） | ⚠️ |
| `goalcli audit goal` | `goalcli score` | ✅ |
| `goalcli bootstrap repo` | `make bootstrap`（待补） | 🔴 |

---

## 退出码（沿用原文 §280）

| Code | 含义 | goalcli 是否兼容 |
|---|---|---|
| 0 | 通过 | ✅ |
| 1 | 业务失败（违规） | ✅ |
| 2 | 用法错误 | ✅ |
| 3 | 配置缺失 | ⚠️（部分） |
| 4 | 环境异常 | ⚠️（部分） |

---

## 下游采用

下游库应消费 `goalcli` 合约与 evidence 命名。可以复制、包装或调用 `goalcli`，但不得把历史旧名作为并列 authority。

**当前阶段**：`goalcli` 是唯一机器裁判与执行器名称。


---

> 本文件被 `scripts/check_docs.sh` 列为 required，删除会阻断 `make docs-check` / `make governance-check` / release pipeline。
