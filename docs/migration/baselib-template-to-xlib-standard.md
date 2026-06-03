# baselib-template 到 xlib-standard 迁移指南

## 目标

把旧 `baselib-template` 身份迁移为 `xlib-standard`，并把旧默认下游 `foundationx` 迁移为 `kernel`。迁移完成后，旧名只允许出现在迁移 ADR、迁移指南、历史变更记录和兼容性说明中。

## 名称规则

| 旧名 | 新名 | 允许保留位置 |
| --- | --- | --- |
| `baselib-template` | `xlib-standard` | ADR、迁移指南、CHANGELOG、兼容性说明 |
| `foundationx` | `kernel` | ADR、迁移指南、历史 compatibility note |

## 迁移要求

- README 主标题和主叙事必须使用 `xlib-standard`。
- 生成示例默认使用 `kernel` / `github.com/ZoneCNH/kernel`。
- 标准文档必须声明 `xlib-standard` 同时承担 Standard Source、Go Reference Template、Generator、Harness 和 Evidence Runtime。
- Go module path、包名、render script 和 CI 迁移由实现 worker 完成；本文档只定义语义约束。

## 验证

- `rg -n "baselib-template|foundationx" README.md docs .agent` 的命中必须能归类为迁移文档语境。
- `GOWORK=off make docs-check` 必须通过。
- Full release 前必须补充 downstream integration Evidence。
