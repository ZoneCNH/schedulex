# 可追踪矩阵

> 收敛规则：每一行标识对应需求的权威产物和必需验证。终态 DONE 证据以最终 release evidence bundle 和新鲜 gate 输出为准；`.agent/*-template.md` 脚手架模板在填充前不是证据。
> 当前治理基线：`docs/standard/goal-runtime.md` 与 `docs/standard/xlib-standard.md` 描述治理目标基线；`docs/project-analysis-20260602.md` v0.3.7 是 2026-06-02 分析快照；当前事实状态由 `.agent/truth-state.yaml`、`.agent/downstream-adoption-status.yaml` 和本矩阵共同收敛。下游同步决策由 `docs/downstream-sync-policy.md` 和 `docs/downstream-matrix.md` 承接。

| REQ | 需求摘要 | 主要产物 | 验证/Evidence | 收敛 owner |
| --- | --- | --- | --- | --- |
| REQ-001 | xlib-standard 身份决策与 README/standard docs 对齐 | README.md; docs/adr/ADR-20260602-001-xlib-standard-role.md; docs/standard/module-boundary.md; docs/standard/repository-roles.md | docs-check; stale-name scan with migration exceptions | docs/runtime |
| REQ-002 | 旧名迁移：baselib-template/foundationx 只在迁移文档语境出现 | docs/migration/baselib-template-to-xlib-standard.md; docs/adr/ADR-20260602-002-kernel-rename.md | rg scan; module/name migration worker evidence | migration/runtime |
| REQ-003 | Core gate 定义 release-final/check/score/preflight | docs/adr/ADR-20260602-003-core-gate.md; .agent/harness.yaml | release-final-check; score gate worker evidence | release-gates |
| REQ-004 | module path/package/render 迁移 | docs/standard/template-generation-contract.md; docs/downstream-matrix.md | render/integration worker evidence | generator/runtime |
| REQ-005 | Full Goal Runtime v3.1 .agent 文件完整，并对齐当前治理基线 | docs/standard/goal-runtime.md; docs/standard/xlib-standard.md; docs/project-analysis-20260602.md; .agent/goal-runtime.md; object-model; state-machine; traceability; harness; evidence/review/release/retro/patch files; .agent/truth-state.yaml; .agent/command-implementation-status.yaml; .agent/release-required-gates.yaml; .agent/evidence-usability.yaml; .agent/downstream-adoption-status.yaml | baseline/delta review; file inventory; docs-check extension | agent-runtime |
| REQ-006 | goalcli/docs-check/score executable gates | .agent/harness.yaml; docs/standard/harness-gates.md | gate implementation worker evidence | release-gates |
| REQ-007 | release manifest/hash/version Evidence | .agent/release-template.md; docs/release.md | release manifest worker evidence | release-evidence |
| REQ-008 | security and secret policy | docs/standard/security-and-secret-policy.md; docs/xgo-integration-boundary.md | security gate; secret leak scan | security/runtime |
| REQ-009 | downstream matrix and kernel integration | docs/downstream-matrix.md; docs/adr/ADR-20260602-002-kernel-rename.md | downstream integration worker evidence | downstream/runtime |
| REQ-010 | x.go integration boundary | docs/xgo-integration-boundary.md; docs/standard/module-boundary.md | boundary gate; import scan | boundary/runtime |
| REQ-011 | 当前事实状态、命令实现状态、release required gates、Evidence 可用性和下游采纳状态可追踪 | .agent/truth-state.yaml; .agent/command-implementation-status.yaml; .agent/release-required-gates.yaml; .agent/evidence-usability.yaml; .agent/downstream-adoption-status.yaml; docs/project-analysis-20260602.md; docs/structural-issues-20260602.md | docs-check; status yaml inventory; traceability review | agent-runtime |
| REQ-012 | P0 debt governance runtime gates and release evidence | .agent/debt/*; cmd/goalcli/debt.go; internal/debtcheck; release/debt/latest.json | debt; debt-evidence; release manifest debt evidence | debt-governance |
