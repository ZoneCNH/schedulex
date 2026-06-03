# 规格

## 目标

对齐独立标准仓库 [`xlib-standard`](https://github.com/ZoneCNH/xlib-standard)，并在同一仓库中提供 Go 参考模板、generator、Harness 和 Evidence Runtime。

## 需求

- 提供可编译 Go 模板、contracts、examples、CI workflow、Harness Gate、Evidence artifact、release 和复盘模板。
- `scripts/render_template.sh` 可以生成 `kernel` 形态并通过 `GOWORK=off go test ./...`。
- 持久同步默认下游为 `kernel`；`corekit` 仅作为中性路径 smoke/integration 验证目标，不作为持久下游同步目标。
- 旧 `baselib-template` / `foundationx` 只保留在迁移文档语境中。
- 禁止隐式读取 `/home/k8s/secrets/env/*`；该路径只属于调用方部署配置。

## 非目标

- 不依赖 `x.go`。
- 不包含 `x.go` 业务模型。
- 不把生成库真实 runtime 塞回标准仓库。

## 目标编号

- 目标：`GOAL-20260602-001`
- 当前基线：`docs/goal.md` v2.9.3 Complete；实现与风险基线参考 `docs/project-analysis-20260602.md` 的 P0/P1/P2 与 52 项问题清单。
