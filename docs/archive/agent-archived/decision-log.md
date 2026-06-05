# 决策日志

| 日期 | 决策 | 理由 | 已拒绝方案 |
| --- | --- | --- | --- |
| 2026-06-02 | `xlib-standard` 作为统一的 standard/template/generator/Harness/Evidence runtime | 消除拆分身份，并让 README、docs 和 `.agent` 与 Full Goal Runtime v3.1 对齐 | 在迁移或历史上下文之外继续把 `baselib-template` 作为主实现仓库 |
| 2026-06-02 | 默认下游为 `kernel` | 建立 L0 集成目标，避免旧 `foundationx` 语义歧义 | 在迁移或历史上下文之外继续把 `foundationx` 作为默认生成库 |
| 2026-06-02 | 完成声明必须同时具备 docs/runtime evidence 和可执行 gate evidence | 防止只有文档的 DONE claim | 把 MVA 或 docs-only 视为最终完成 |
