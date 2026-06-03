# Full Goal Runtime v3.1

Runtime: `xlib-standard` Full Goal Runtime v3.1.

## 目的

把 Goal、Spec、Requirement、Acceptance Criteria、Design、ADR、Plan、Task、Test、Evidence、Risk、Decision、Review、Release、Retrospective 和 Patch 连接为同一条可审计链。完成不是 MVA，而是 `DONE with evidence:`：REQ-001..REQ-010 全部关闭、release-final-check 通过、score >= 9.8、kernel downstream integration 通过、manifest 完整。

## 必需工件

- `.agent/object-model.md`
- `.agent/state-machine.md`
- `.agent/traceability-matrix.md`
- `.agent/harness.yaml`
- `.agent/evidence-protocol.md`
- `.agent/review-template.md`
- `.agent/release-template.md`
- `.agent/retrospective-template.md`
- `.agent/risk-register.md`
- `.agent/decision-log.md`
- `.agent/rollback-protocol.md`
- `.agent/prompt-patches.md`
- `.agent/harness-patches.md`
- `.agent/rule-patches.md`
- `.agent/truth-state.yaml`
- `.agent/command-implementation-status.yaml`
- `.agent/release-required-gates.yaml`
- `.agent/evidence-usability.yaml`
- `.agent/downstream-adoption-status.yaml`

## 停止条件

只有当 Evidence 证明当前 slice 完成，或 blocker 已记录 owner、scope 和 next action 时才停止。最终 goal claim 必须包含命令、输出、manifest/checksum 状态、score 状态和 known gaps。
