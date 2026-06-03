# goalcli v0.1.0 roadmap

Status: execution roadmap。normative authority 位于 `docs/standard/goalcli-runtime.md` 和 `docs/adr/ADR-20260603-001-goalcli-runtime.md`。

## Authority map

- Root proposal: `.worktree/goalcli-v0.1.0-plan.md`。
- Context brief: `.omx/context/goalcli-v0.1.0-team-20260603T005302Z.md`。
- Runtime standard: `docs/standard/goalcli-runtime.md`。
- Runtime ADR: `docs/adr/ADR-20260603-001-goalcli-runtime.md`。
- Source evidence ledger: `.agent/evidence/ledger.jsonl`。
- 当前完成状态: `mva_status: complete`。

## PR alias table

| Root plan area | Roadmap alias | Current v0.1.0 state |
| --- | --- | --- |
| PR-0/1/2 authority and registry setup | PR-1 authority split | 已记录并调和到 MVA completion |
| PR-3/4/5 Harness and command wiring | PR-4 command-backed slice | 已实现为 G12-G16 MVA-blocking checks |
| PR-6/8 evidence and final checks | PR-6/PR-8 evidence reconciliation | 是 full MVA evidence 的 required step |
| PR-9+ CLI/product expansion | v0.2.0+ | 不属于 v0.1.0 scope |

## MVA rule

v0.1.0 MVA 暴露 `goal-acceptance`、`goal-delivery`、`goal-handover`、`goal-downstream-adoption`、`goal-certify` 和 `goal-runtime-final`。这些 commands 是 MVA-blocking evidence checks；只有当所有 prerequisite checks 都在 source ledger `.agent/evidence/ledger.jsonl` 中以同一 `GOAL_ID` 调和后，`goal-runtime-final` 才是 completion rollup。
