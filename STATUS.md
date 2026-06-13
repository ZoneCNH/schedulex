# STATUS

| 模块 | 版本 | 状态 | 规模 | 摘要 |
| --- | --- | --- | --- | --- |
| [schedulex](https://github.com/ZoneCNH/schedulex) | v1.0.0 | █████ 发布门禁通过 | 约 398KB / 25 项 | cron/interval/delay 调度、Overlap/Misfire 策略、Locker 扩展点、Clock 注入、示例与 contracts；`release-check`、score、contracts、race、vet、lint、govulncheck 已通过，待提交、合入 `main`、tag 与 GitHub Release 发布。 |

## 发布判定

`schedulex` 可以发布 `v1.0.0`。release 分支最终 gate 已通过；正式发布完成条件是提交已推送并合入 `main`、`v1.0.0` tag 已推送、GitHub Release 已创建，并在 tag 对外可见后补跑网络下游 smoke。

## 当前状态

- 版本锚点：`Makefile`、`pkg/schedulex.Version`、contracts、release manifest schema、downstream fixture 均对齐 `v1.0.0`。
- 当前分支：`release/v1.0.0`。
- 当前验证：`GOWORK=off make release-check VERSION=v1.0.0` 通过；score `10.0`，覆盖率 `98.2%`。
- 当前工作树：发布改动与新建 `STATUS.md` 已验证，待提交。
- OMX team：分析任务已完成，待发布收口后关闭 team runtime。

## 待完成项

- 提交并推送 `release/v1.0.0`，再快进合并到 `main`。
- 推送 `v1.0.0` tag 并创建 GitHub Release。
- tag 可见后运行 `SCHEDULEX_DOWNSTREAM_NETWORK=1 VERSION=v1.0.0 ./scripts/check_downstream_smoke.sh`。
- 更新上层索引仓库 `/home/ZoneCNH/STATUS.md` 的 schedulex 行。
