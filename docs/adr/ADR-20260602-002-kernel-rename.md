# ADR-20260602-002: 默认下游从 foundationx 迁移到 kernel

## 状态

Accepted — 2026-06-02

## 背景

旧示例名 `foundationx` 在 README、生成文档和仓库角色中承担默认下游示例，但 Full Goal Runtime v3.1 要求默认下游集成目标为 `kernel`，并把 `configx`、`observex`、`testkitx` 与 profile 库放入明确矩阵。

## 决策

默认下游名迁移为 `kernel`。`kernel` 是 L0 通用基础能力库，承载 context、error、config、logging、metrics、lifecycle、health 和 test helper 等 primitive。旧 `foundationx` 只作为迁移文档语境出现。

## 影响

- README 的生成示例使用 `kernel`。
- [下游矩阵](../downstream-matrix.md) 以 `kernel` 为 L0 行。
- `docs/standard/repository-roles.md` 和 `docs/standard/layering.md` 使用 `kernel` 而不是 `foundationx` 作为当前默认下游。
- module/name 迁移、render 集成和真实 downstream verification 由对应 worker slice 完成。

## Evidence

- README 生成 `kernel` 示例。
- `docs/downstream-matrix.md` 包含 `kernel` 行。
- `.agent/traceability-matrix.md` REQ-002/REQ-009 追踪项。
