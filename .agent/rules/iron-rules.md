# Iron Rules — Goal Runtime 第一性铁律

> SSOT. 本文件由 [`docs/adr/ADR-20260603-002-rules-registry.md`](../../docs/adr/ADR-20260603-002-rules-registry.md) 锁定。
> 任何分歧以本文件 + [`registry.yaml`](./registry.yaml) 为准；`.worktree/goal-patch.md` 仅作历史推导记录。
>
> **叙事/解释层**：完整背景、9 层架构、v0.1.0 五主线见 [`.agent/standard/goal-runtime-canonical.md`](../standard/goal-runtime-canonical.md)（PR #30 引入）。本文件是机器消费层；canonical 把 RULE-EVIDENCE-001 单独列为第 8 条以利叙事，本文件把它合入第 1 条以利归一化——两者通过下方"七律"的括注 RULE-* 编号保持稳定映射。

铁律是 [`registry.yaml`](./registry.yaml) 中 119 条 P0 规则的归一化压缩，**违反任何一条都必须阻断 Release**。

## 七律

1. **没有 Evidence，不允许 DONE** —— Task/Issue/Goal/Release 不能只靠描述完成 (RULE-CORE-001 / RULE-EVIDENCE-001 / RULE-DOD-001)。
2. **不恢复上下文，不允许设计** —— 任何 Goal 必须先扫描仓库、文档、CI、已有规则 (RULE-CORE-002 / RULE-CONTEXT-001)。
3. **没有 Acceptance Criteria 的需求，不允许实现** —— Requirement → AC → Test → Evidence 不能断链 (RULE-CORE-003 / RULE-SPEC-003)。
4. **所有变更必须可追踪** —— Goal → Req → AC → Task → Issue → Commit → PR → Evidence → Release 闭环 (RULE-CORE-004 / RULE-TRACE-001 / RULE-TRACE-002)。
5. **main 只做基线，所有开发必须 worktree** —— 禁止 main 直接 push、直接合未通过 Gate 的 PR；同时禁止 secret 进入代码/文档/Evidence/Release (RULE-WORKTREE-001 / RULE-MAIN-SYNC-001 / RULE-MERGE-001 / RULE-SECRET-001)。
6. **Harness Gate 是机器裁判，失败必须阻断** —— P0 Gate 不可豁免；本地 Gate 与 CI Gate 必须一致 (RULE-CORE-005 / RULE-HARNESS-003 / RULE-WAIVER-002 / RULE-GATE-CONSISTENCY-001)。
7. **重复问题必须升级为 Rule / Harness / Prompt Patch** —— Retrospective 不能只是总结，必须产出可执行 Patch (RULE-CORE-006 / RULE-RETRO-003 / RULE-SI-001)。

## 标准退出码

`goalcli` 与所有 Gate 命令统一遵守，便于 Makefile / Hooks / CI / Agent 串接：

| 退出码 | 含义 | 触发规则 |
|---|---|---|
| 0 | OK | — |
| 1 | 通用失败 | — |
| 2 | 参数错误 | — |
| 5 | worktree / main 违规 | RULE-WORKTREE-* / RULE-MAIN-SYNC-* / RULE-BRANCH-* | ✅ `goalcli worktree-guard` / `main-guard` |
| 6 | schema 校验失败 | RULE-SCHEMA-* | ✅ `goalcli policy-schema` |
| 7 | secret / 凭据泄漏 | RULE-SECURITY-* / RULE-SECRET-* | ✅ `goalcli secrets` |
| 8 | Evidence 缺失或伪造 | RULE-EVIDENCE-* / RULE-CORE-001 | ✅ `goalcli evidence-check` |
| 9 | Traceability 断链 | RULE-TRACE-* / RULE-CORE-004 | ⚠️ **GAP**：`traceability-check` 尚未实现 |
| 10 | Release 不完整 | RULE-RELEASE-* / RULE-REL-ARTIFACT-* | ✅ `goalcli release-evidence-check` / `release-final-check` |

## 已知 P0 Gap

> 本节是诚实性披露：以下规则虽属 P0，但 `enforced_by` 尚未对应任何已实现的命令；
> 在 [`registry.yaml`](./registry.yaml) 中标记为 `status: indexed`，等待落地。

- **Traceability Gate**：RULE-CORE-004, RULE-TRACE-001, RULE-TRACE-002, RULE-TRACE-ALG-001, RULE-TRACE-ALG-002（共 5 条）。需要 `goalcli traceability-check` 子命令，按 `.agent/traceability-matrix.yaml` 校验 Goal → Req → AC → Task → Issue → Commit → PR → Evidence → Release 链路完整性，断链返回退出码 9。
- **Self-improving Gate**：RULE-CORE-006, RULE-RETRO-*, RULE-SI-* 等。需要 Retrospective/Patch 校验命令。

P0 gap 的状态可通过 `make rules-verify` 持续检测（任何 active 规则若引用不存在的命令将阻断 CI）。

## 不在铁律中的内容

非 P0 规则（治理、文档、自动化、度量等共 300 条）登记在 [`registry.yaml`](./registry.yaml)，按需启用。
具体分类规则见各 `*-rules.md`：[goal](./goal-rules.md) · [worktree](./worktree-rules.md) · [commit](./commit-rules.md) · [pr](./pr-rules.md) · [evidence](./evidence-rules.md) · [release](./release-rules.md) · [harness](./harness-rules.md) · [security](./security-rules.md) · [issue](./issue-rules.md) · [risk-decision](./risk-decision-rules.md) · [self-improving](./self-improving-rules.md)。
