# 变更日志

## 未发布

暂无。

## v1.0.0 - 2026-06-13

### 发布

- 发布 `github.com/ZoneCNH/schedulex` L1 deterministic scheduler 稳定版。
- 锁定公共 API：trigger、`Clock`、`Locker`、misfire/overlap 策略、事件输出、snapshot 和 scheduler 生命周期。
- 将版本、contracts、release manifest schema、downstream fixture、README、standard docs 和 release docs 对齐到 `v1.0.0`。
- 保持生产代码仅依赖 Go 标准库；下游锁实现通过 `Locker` 接口注入。

### 验证

- 发布前 gate：unit、vet、race、boundary、contracts、docs-check、score、release-final-check、release-preflight。

## v0.4.6 - 2026-06-04

### Changed

- 将 `Clock` 从 `Trigger.Next` 签名移到 `SchedulerOptions.Clock`，恢复 trigger 仅按传入时间计算下一次触发时间的纯函数语义。
- 文档与 contract fixture 同步更新 clock 注入边界。

### Fixed

- `Start` 现在会在失败时等待调度循环退出，避免锁获取失败路径泄漏 goroutine。

## v0.4.5 - 2026-06-04

### Fixed

- `Start` 在 scheduler 已取消后返回 `ErrSchedulerShutdown`，避免重启已关闭实例时产生不确定状态。
- `NewScheduler` 现在复制 `SchedulerOptions`，创建后修改原始 options 不会影响 scheduler 行为。
- `Scheduler.Register` 会拒绝已经取消的 job context，避免注册永远不会执行的 job。
- `JobEvent` 改为深拷贝 `JobID`、`Err` 和 `Attempt` 标签，避免后续修改输入对象影响已发布事件。
- Snapshot/事件时间戳统一走注入的 `Clock`，提升测试可控性。
- `Scheduler.Stop` 同时等待 ready 但尚未执行的 jobs，确保 stop 返回后不再有 job 执行。
- `Scheduler.Reset` 在清空 job map 后异步取消旧 jobs，避免用户回调在 scheduler 锁内执行。
- `Register` 默认初始化 `RetryBackoff`，让手动构造的 `JobSpec` 与 helper 构造结果一致。
- `JobSnapshot` 拷贝 `RunningSince` 指针，避免快照使用方间接修改内部状态。
- `Register` 会复制 `Attempts` map，避免调用方后续修改影响 scheduler 状态。
- `IntervalTrigger` 与 `CronTrigger` 拷贝输入 `time.Location`，避免外部指针修改影响触发器。

## v0.4.4 - 2026-06-04

### Changed

- 发布可复用 `RunReleaseReadiness`，统一 release readiness gate 的执行、结构化结果与 CLI 输出。
- `scripts/check_schedulex_release.sh` 改为调用 `go run ./cmd/schedulex-release-check`，不再维护重复 bash 检查逻辑。
- 新增 `docs/release.md` 说明发布门禁命令和验证边界。
- `make release-check` 现在执行结构化 release readiness CLI。

### Fixed

- 补齐 release readiness 对 race、coverage、artifact、API、contract、boundary、score、downstream、documentation、schema、git state 和 tag readiness 的可测试诊断。
- 修复 `make release-check` 中 GOWORK 环境变量传递位置。

## v0.4.3 - 2026-06-04

### Changed

- 将生产代码收敛为零第三方依赖，`go.mod` 仅保留 module 与 Go 版本声明。
- 新增 `test/downstream-smoke` 最小下游工程，验证 `go test github.com/ZoneCNH/schedulex/test/downstream-smoke` 可作为消费者 smoke。
- 下游 fixture 迁移到 `github.com/ZoneCNH/schedulex/release/downstream-adoption/fixture` 并使用 `v0.4.3`。
- release 脚本与文档同步去除 `github.com/robfig/cron/v3` 生产依赖声明。
- `CronTrigger` 改为项目内解析实现，支持六字段 cron、`*`、`*/n` 与数字字段。

### Fixed

- 修复 first-party downstream fixture 旧 module path 导致的 `go mod download` 失败。
- `check_downstream_smoke.sh` 增加 GitHub URL 可达性检查，并在网络不可用时明确跳过远端发布验证。
- 为 Go 1.20 下 `time.WithoutCancel` 不可用提供兼容实现。

## v0.4.2 - 2026-06-04

### Added

- 引入 `scripts/check_api_snapshot.sh` 与 `make api-check`，将 `docs/api.md`、`contracts/public_api.snapshot` 与实际公开 API 锁定为同一份 truth set。
- 新增最小下游 adoption fixture 与 smoke 脚本，覆盖 `go mod download github.com/ZoneCNH/schedulex@v0.4.2`、零第三方生产依赖检查和 main path import。
- README、docs/spec、docs/test-strategy 和 docs/release 补齐 API snapshot、downstream adoption 与 release readiness 说明。

### Changed

- `make release-final-check` 现在串联 docs、contracts、API snapshot、score、race、vet、release readiness 和下游 smoke。
- `make release-preflight VERSION=vX.Y.Z` 增加 tag readiness 校验，要求工作树 clean 且本地/远端 tag 尚不存在。
- Contract fixture 记录 API snapshot、downstream adoption、module path 与生产依赖边界。
- `scripts/check_schedulex_release.sh` 汇总 release readiness，并输出 JSON 与文本报告。

### Fixed

- `Scheduler.Stop` 现在等待 running jobs 完成，避免 stop 返回后仍可能更新 job state。
- 分布式锁失败事件携带 `ErrLockUnavailable`，让下游可区分 lock contention 与 job handler 错误。

## v0.4.1 - 2026-06-04

### Changed

- 将 downstream fixture 的 module path 从 `github.com/ZoneCNH/schedulex/downstream/smoke` 改为 `github.com/ZoneCNH/schedulex/release/downstream-adoption/fixture`，避免与仓库内已有 package 路径冲突。
- release readiness 现在检查 contracts/schema、API snapshot、生产依赖、downstream fixture、score report、Makefile target、文档和 git/tag 状态。
- README、docs/release、docs/spec 和 docs/test-strategy 同步补齐 release readiness gate 与 downsteam smoke 说明。

## v0.4.0 - 2026-06-04

### Changed

- 将公共 API 从泛型 payload/channel runner 收敛为确定性 scheduler：`JobSpec`、trigger、misfire/overlap policy、snapshot、Clock、EventSink 与 `Scheduler` 生命周期。
- 删除 `internal/queue`、`internal/retry`、`pkg/fx`、`pkg/worker` 旧 runtime，避免遗留 runner 语义污染 L1 scheduler 边界。
- 重写 examples 为 cron、delay、interval、misfire、overlap、distributed lock、clock injection 与 resilient job 八个最小示例。
- README、docs/spec、docs/api、docs/test-strategy、contracts 与 release manifest schema 对齐 scheduler API 与语义。

### Removed

- 移除 `fx` 与 worker API；下游应迁移到 scheduler jobs、trigger 和 EventSink。

## v0.1.1 - 2026-06-01

### Changed

- 标准化所有公开文档与 release metadata 中的仓库身份为 `schedulex` / `github.com/ZoneCNH/schedulex`，移除 `baselib-template`、`foundationx`、`Templatex`、`Goalcli` 等模板/旧仓库命名。
- 同步更新 quality score、contract fixture、README、docs 与 release manifest schema，使当前公共身份、module path、artifact 名称与版本锚点一致。
- 保留迁移检测逻辑中针对旧命名的扫描规则，确保未来不会重新引入模板残留。

## v0.1.0 - 2026-06-01

### Added

- 发布 `github.com/ZoneCNH/schedulex` 的 L1 deterministic scheduler 最小可用版本。
- 支持 `DelayTrigger`、`IntervalTrigger` 与 `CronTrigger`，覆盖 delay / interval / cron 三类调度。
- 提供 `JobSpec`、`Scheduler`、`JobEvent`、`EventSink`、`Clock` 与 `Locker` 等稳定 API。
- 实现 `OverlapPolicy`、`MisfirePolicy`、retry backoff、snapshot、job lifecycle 与 graceful stop/reset。
- 提供 clock injection、distributed lock、misfire、overlap、resilient job 等示例。
- 建立 contracts fixture、release manifest schema、score report 与 docs/test evidence。
