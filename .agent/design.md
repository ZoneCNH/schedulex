# 设计

## 结构

- `pkg/templatex` 是可编译 Go 参考模板的公共 API 占位包。
- `internal/` 包含模板自检、release manifest、CLI gate 等仓库内部工具。
- `scripts/` 包含文档、边界、contract、secret、integration 和 release gate。
- `.github/workflows/` 承载 CI、集成、安全和发布检查。
- `docs/standard/`、`docs/adr/`、`docs/downstream-matrix.md` 和 `docs/xgo-integration-boundary.md` 是标准、身份和边界事实源。
- `.agent/` 是 Full Goal Runtime v3.1 的目标、对象、状态机、traceability、Harness、Evidence、review、release、retrospective 和 patch runtime。
- `release/manifest/` 包含发布 Evidence 模板；`latest.json` 与 `.sha256` 是生成产物，不提交源码历史。
