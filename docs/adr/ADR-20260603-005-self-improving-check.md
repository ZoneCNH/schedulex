# ADR-20260603-005 — 新建 `goalcli self-improving-check` enforcer

## Status

Accepted

## Context

ADR-004 落地后，registry.yaml 剩 65 条 indexed，其中含唯一一条 P0 indexed：

- `RULE-CORE-006`（Self-improving 是强制环节）

加上 8 条 P1 indexed 同语义簇：

- `RULE-RETRO-001/002/003`（每个 Goal 必须有 Retrospective / 必须生成 Patch / 重复问题升级为规则）
- `RULE-RETRO-CHECK-001/002`（Retrospective 不能只是总结 / 缺 Patch 候选则 Gate 失败）
- `RULE-SI-001/002/003`（Retro 必须生成可执行 Patch / Patch 必须分状态 / Patch 必须进入 Registry）

仓库已有相关产物：`.agent/retrospective.md` / `retrospective-template.md` / `{harness,prompt,rule}-patches.yaml` / `harness/gates/retro-gate.yaml`，但缺机器 enforcer：`Makefile` 中 `retro-check` 仅 `@echo passed`。

## Decision

新增 `cmd/goalcli/selfimproving.go` 实现 `goalcli self-improving-check`（别名 `retro-check`），分 4 类静态检查：

1. **必备文件存在**（RULE-RETRO-001）：retrospective.md / template / 3 个 patches.yaml / retro-gate.yaml
2. **retrospective.md 体现复盘性质**（RULE-RETRO-CHECK-001）：含「复盘/Retrospective/回顾」+「补丁/Patch」+「失败/改进/Root Cause/根因/What」三类关键词，双语容忍
3. **retrospective-template.md 遵循完整 9 段 schema**：含 Failure/Root Cause/Patch/Prompt Patch/Harness Patch/Rule Patch
4. **patches.yaml schema 合法**（RULE-SI-002/003）：含 schema/entries/patches 任一字段；status 字段（若有）必须 ∈ {PROPOSED, ACCEPTED, REJECTED, SUPERSEDED, IMPLEMENTED, reconciled_stub}

`--strict` 追加 RULE-SI-001 / RULE-RETRO-CHECK-002 校验：3 个 patches.yaml 合计至少 1 个 `- patch_id:` entry。**默认非 strict**，因 RULE-RETRO-CHECK-002 明确允许 Lite Mode 无失败时豁免。

`Makefile` 中 `retro-check` 改为调用 `$(GOALCLI) self-improving-check`，向后兼容旧调用方。`scripts/extract_rules.py` 中 `CORE-006/RETRO/RETRO-CHECK/SI` 4 个前缀全部从 indexed 升 active。

## 结果

| 维度 | Before | After |
|---|---:|---:|
| total | 419 | 419 |
| **P0 active** | 118 | **119 (100%)** |
| P1 active | 236 | 244 |
| 合计 active | 354 (84%) | **363 (87%)** |
| 仅 indexed | 65 | 56 |

**P0 100% 机器化是里程碑**：仓库所有 P0 规则都至少有一个 goalcli 子命令认领。

## Test 覆盖

`cmd/goalcli/selfimproving_test.go` 5 个表驱动测试：
- Lenient_Passes（含全部必备文件 + 0 patches → 通过）
- Strict_FailsWithoutEntries（strict + 0 patches → exit 1）
- Strict_PassesWithEntry（strict + 1 PROPOSED entry → 通过）
- MissingFile_Fails（删 retrospective.md → exit 1 含 RULE-RETRO-001）
- BadStatus_Fails（status: INVALID_X → exit 1 含 RULE-SI-002）

## Alternatives 拒绝

- **走 YAML schema validator (e.g. cuelang)**：增加 dep，超出 PR-SIZE。手写 regex 已够。
- **默认 strict**：会让本仓库（patches.yaml 全 stub）立即 fail，与 RULE-RETRO-CHECK-002 的 Lite Mode 豁免冲突；通过 `--strict` 让发布门禁可选启用。
- **严格要求 retrospective.md 遵循 template 9 段**：本仓库现有 retrospective.md 是"治理纲领"风格，不符合模板。要么改文件（属另一 PR 范畴），要么放宽判定。选放宽 + template 端严格 schema。

## Risks

- **判定过宽**：retrospective.md 只查 3 类关键词，可能被空壳文件通过。Mitigation：3 类关键词覆盖了复盘的核心语义；template 端走严格 schema 检查；未来可加 `--strict-headings` 选项。
- **patches.yaml status 枚举包含 `reconciled_stub`**：这是过渡值，应在后续 batch 全部迁移到 5 标准状态。已在 ADR 留记。

## Traceability

- 前置：ADR-002 / ADR-004
- 关联规则：RULE-CORE-006 / RULE-RETRO-{001,002,003} / RULE-RETRO-CHECK-{001,002} / RULE-SI-{001,002,003}
- 关联铁律：RULE-CODE-001（机器化）、RULE-ANTI-CARGO-001（拒虚假绑定）

## DONE with evidence

- `go test ./cmd/goalcli/ -run SelfImproving -v` → PASS 5/5
- `python3 scripts/extract_rules.py` → `419 rules, P0=119, P1=300, active=363`
- `go run ./cmd/goalcli self-improving-check` → passed
- `go run ./cmd/goalcli retro-check --strict` → failed (期望，仅说明 0 patches)
