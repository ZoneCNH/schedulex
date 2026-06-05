# GOAL-20260602-001

将 `xlib-standard` 升级为 Full Goal Runtime v3.1 的统一标准、Go 参考模板、Generator、Harness 和 Evidence Runtime 仓库。

## 成功条件

- `README.md`、`docs/standard/`、ADR 和 `.agent/` 不再把旧 `baselib-template` / `foundationx` 当作主身份或默认下游。
- 旧名只用于迁移文档语境。
- 默认下游集成目标为 `kernel`；目标生成库矩阵覆盖 `configx`、`observex`、`testkitx`、`postgresx`、`redisx`、`kafkax`、`natsx`、`taosx`、`ossx`、`clickhousex`。
- `/home/k8s/secrets/env/*` 只属于调用方部署路径，不能被 `xlib-standard`、`kernel` 或生成库读取、提交、记录到日志或写入 Evidence/manifest/PR。
- 最终发布证据必须使用 `DONE with evidence:`，并通过 `GOWORK=off make release-final-check`、`GOWORK=off make release-preflight VERSION=<version>` 和 `goalcli score --min 9.8`。
