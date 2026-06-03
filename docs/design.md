# 设计

`xlib-standard` 是 [`https://github.com/ZoneCNH/xlib-standard`](https://github.com/ZoneCNH/xlib-standard) 的统一标准、Go 参考模板、generator、Harness 和 Evidence Runtime 仓库。生成的库是独立 Go module。公共 API 位于 `pkg/{{PACKAGE_NAME}}`，内部辅助代码位于 `internal/`，contracts 位于 `contracts/`，运行 Evidence 位于 `release/manifest/`。`scripts/render_template.sh` 是模板到具体基础库的唯一内置渲染入口。

## 目录职责

- `docs/standard/`：标准、分层、边界、DoD、Evidence 和 release 规则。
- `docs/adr/`：身份、kernel 默认下游和 core gate 的决策记录。
- `.agent/`：Full Goal Runtime v3.1 的对象、状态、traceability、Harness、Evidence、review、release、retrospective 和 patch runtime。
- `pkg/templatex`：可编译 Go 参考模板公共包。
- `internal/`：模板自检、release manifest 和 gate 内部工具。
- `scripts/`：文档、边界、contract、secret、integration 和 release gate。

## 下游

默认下游集成目标是 `kernel`，中性路径 smoke 目标是 `corekit`。旧 `baselib-template` / `foundationx` 名称仅作为迁移文档语境保留。

## 发布

发布前必须通过 Harness Gate，并确认入口文档仍引用 [`https://github.com/ZoneCNH/xlib-standard`](https://github.com/ZoneCNH/xlib-standard)，然后生成 `release/manifest/latest.json` 与 `release/manifest/latest.json.sha256`。这些文件是 release Evidence artifact，不提交到源码历史。`make release-check` 会先运行 CI 和 integration gate，再以 `CHECK_STATUS=passed` 生成 manifest。最终发布使用 `GOWORK=off make release-final-check`、`GOWORK=off make release-preflight VERSION=<version>` 和 `goalcli score --min 9.8`。
