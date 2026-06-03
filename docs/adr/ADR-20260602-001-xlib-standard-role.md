# ADR-20260602-001: xlib-standard 合并标准、模板、Generator、Harness 与 Evidence Runtime

## 状态

Accepted — 2026-06-02

## 背景

Full Goal Runtime v3.1 要求一个仓库同时提供标准文本、可编译 Go 参考模板、生成器、Harness gate 和 Evidence runtime。旧叙事把 `xlib-standard` 视为外部标准源、把 `baselib-template` 视为实现仓库，导致 README、docs、.agent 和 release Evidence 出现身份漂移。

## 决策

`xlib-standard` 是唯一主身份，承担五类职责：

1. Standard Source。
2. Go Reference Template。
3. Generator。
4. Harness。
5. Evidence Runtime。

旧名 `baselib-template` 只保留在迁移 ADR、迁移指南、历史变更记录和兼容性说明中，不得作为新 README 主标题、生成默认值或 release completion 主体出现。

## 影响

- README 必须明确五类职责。
- `docs/standard/module-boundary.md` 不再禁止 `xlib-standard` 拥有 generator、Harness 或 Evidence 实现。
- `docs/standard/repository-roles.md` 不再把 `baselib-template` 作为主实现仓库。
- `.agent/` 必须描述 Full Goal Runtime v3.1 的状态机、对象模型、traceability、Evidence、release、rollback 和 patch protocol。

## Evidence

- README 五类职责章节。
- `docs/standard/module-boundary.md` 允许内容与禁止内容。
- `docs/standard/repository-roles.md` 角色表。
- `.agent/traceability-matrix.md` 中 REQ-001 追踪项。
