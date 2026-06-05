# Goal Runtime Canonical Standard v1.0

> 本文档是 xlib-standard 的 Goal Runtime **唯一权威规格**。  
> 原始演进合集见 `.agent/inbox/goal-patch-v1.0-to-v2.2.md`（只读归档，13856 行）。  
> 本文档为裁决版，目标 ≤ 800 行，只保留经过审批的、有机器化等价实现的规则。

---

## 0. 适用范围

| 对象 | 是否适用 |
|---|---|
| `xlib-standard` 本仓库 | ✅ |
| 下游库（kernel / configx / observex / ...） | ✅（通过 Adoption Manifest） |
| 应用层（x.go 等） | 仅作为消费方约束 |

---

## 1. 八条铁律（不可违反，违反必阻断）

> **机器消费层**：见 [`.agent/rules/iron-rules.md`](../rules/iron-rules.md) + [`.agent/rules/registry.yaml`](../rules/registry.yaml)（PR #34 引入）。本节为叙事/解释层，两者交叉互证：iron-rules 把 RULE-EVIDENCE-001 并入第 1 条，本节单独列为第 8 条，编号映射在每行的"机器化实现"列保持稳定。

| ID | 铁律 | 机器化实现 |
|---|---|---|
| RULE-CORE-001 | 没有 Evidence 不允许 DONE | `goalcli evidence-check` / `make evidence-check` |
| RULE-CORE-002 | 必须从真实上下文开始 | `goalcli context-profile-check` |
| RULE-CORE-003 | 需求必须可验证（Req→AC→Test→Evidence） | `.agent/acceptance-matrix.yaml` + `goalcli acceptance-matrix` |
| RULE-CORE-004 | 所有变更必须可追踪 | `.agent/traceability-matrix.yaml` + `goalcli trace-check`（待实现） |
| RULE-CORE-005 | Harness 是机器裁判 | `cmd/goalcli/` + `make ci` |
| RULE-CORE-006 | Self-improving 强制 | `goalcli retro-check`（每 Goal 必须有 Retro+Patch 候选） |
| RULE-WORKTREE-001 | 禁止 main 开发 | `.githooks/pre-commit` + `pre-push` + GHA `worktree-guard` + GitHub branch protection（四道防线） |
| RULE-SECRET-001 | 禁止 secret 进入代码/文档/Evidence/Release | `scripts/check_secrets.sh` + `.githooks/pre-commit` + GHA `security.yml` |

**铁律之外的所有规则**：参考原文，按需采纳，不作强制。

---

## 2. 九层架构（参考映射，**不强制目录重构**）

| 层 | 文档定义 | 本仓库实际位置 |
|---|---|---|
| L0 Constitution | `.agent/constitution/` | `CONSTITUTION.md`（仓库根） |
| L1 Rules | `.agent/rules/` | `.agent/rules/` ✅ |
| L2 Policies | `.agent/policies/` | 散落于 `.agent/*.yaml`（不重构） |
| L3 Schemas | `.agent/schemas/` | 散落 yaml + Go struct（不重构） |
| L4 Harness Gates | `.agent/harness/` | `cmd/goalcli/` + `scripts/harness/` |
| L5 Registries | `.agent/registries/` | `.agent/*-registry.yaml`（命名后缀代替子目录） |
| L6 Goal Packs | `.agent/goals/<GOAL-ID>/` | 暂未启用（按需） |
| L7 Automation | goalcli / GHA | `goalcli` + `.github/workflows/` |
| L8 Evidence & Audit | `.agent/evidence/` + `release/evidence/` | ✅ 已对齐 |
| L9 Self-improving | Patch Registry | `.agent/retrospective-*.yaml` |

**裁决**：物理重构成本远大于收益，沿用现有平铺结构 + 命名后缀代替目录层级。

---

## 3. v0.1.0 最小可执行目标

只完成两件事：

### Track A：闭合 RULE-WORKTREE-001 的本地防线
- `git config core.hooksPath .githooks`（纳入 `make doctor` / `make bootstrap`）
- `worktree-guard.yml` 增加 `push: branches: [main]` 触发
- 文档化 GitHub branch protection 必备设置

### Track B：闭合 RULE-SECRET-001 的提前拦截
- `.githooks/pre-commit` 增加调用 `scripts/check_secrets.sh`
- 保留 GHA 二次拦截不变

**其他 Track（控制平面重构、第二套执行内核、Trust Score 等）一律延后到有真实痛点时再启动。**

---

## 4. 命名规约：goalcli

详见 `.agent/standard/goalcli-mapping.md`。
要点：`goalcli` 同时是标准合约名、机器执行面和本仓库实现入口。
历史旧名不再作为并列 authority。

---

## 5. 反过度工程化清单（明确不做）

- ❌ `.agent/` 90 文件物理迁移到九层目录
- ❌ 新建第二套 `tools/goalcli/` Go 内核（`cmd/goalcli` 已覆盖）
- ❌ Rule Compiler / Trust Score / Drift Budget / Agent Lease 等高级特性
- ❌ 把原文 300+ 条 RULE 全部 lift 进 `.agent/rules/`
- ❌ 重命名 82 个 Makefile target

---

## 6. 规则增补协议

新规则进入本 canonical，必须满足：
1. 有明确真实痛点来源（不是"理论应该有"）
2. 有机器化检查方式（落到 goalcli 子命令或 hook 或 GHA）
3. 有违反样例 fixture
4. PR 描述说明"为什么之前不需要、现在为什么需要"

否则只能进 `.agent/inbox/` 作为思想储备。

---

## 7. 评分（沿用现有）

`goalcli score --min 9.8` 不变。


---

> 本文件被 `scripts/check_docs.sh` 列为 required，删除会阻断 `make docs-check` / `make governance-check` / release pipeline。
