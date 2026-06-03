# goalcli v0.1.0 x xlib-standard 执行边界方案

生成时间：2026-06-03
更新日期：2026-06-04
目标仓库：`github.com/ZoneCNH/xlib-standard`
目标版本：`v0.4.6`
Goal：`GOAL-20260603-XLIB-GOALCLI-001`

本文是 source-only 执行边界记录，用于说明 `goalcli v0.1.0` 在
`xlib-standard` 中的最小可验收交付面。本文不是发布证据，也不是替代标准文件；
发布证据以 `release/evidence/goalcli/` 中的 JSON 记录为准。

## 1. 目标

把 `xlib-standard` 收敛为由 `goalcli` 驱动的基础库标准工厂，形成可执行的
Goal Runtime、Harness gate、Evidence ledger 和发布验证链。

完成声明只能使用：

```text
DONE with evidence:
```

## 2. Source-only 权威路径

当前执行边界由以下源码文件共同描述：

```text
docs/adr/ADR-20260603-001-goalcli-runtime.md
docs/standard/goalcli-runtime.md
docs/standard/goalcli-cli-contract.md
.agent/harness.yaml
.agent/command-registry.yaml
.agent/makefile-target-registry.yaml
.agent/makefile-baseline.yaml
Makefile
cmd/goalcli/
internal/goalruntime/
release/evidence/goalcli/GOAL-20260603-XLIB-GOALCLI-001.json
```

其中 `cmd/goalcli` 是唯一机器执行面。Makefile、CI 和文档中的 gate 调用必须委托
到 `goalcli`，不能再新增第二套执行入口。

## 3. 运行时分层

```text
Goal Kernel
  -> Harness Runtime
  -> goalcli Executors
  -> Gate Results
  -> Evidence Ledger
  -> Completion / Release / Ecosystem / Governance Extensions
```

最小 Goal Kernel 只表达 `Goal`、`Spec`、`Design`、`Plan`、`Task`、`Test`、
`Evidence`、`Review`。Release、Publishing、Conformance、Ecosystem、
Automation 和 Observability 均作为扩展层进入，不进入 Kernel 本体。

## 4. MVA gate

`goalcli v0.1.0` 的最小验收 gate 为：

```text
goal-acceptance
goal-delivery
goal-handover
goal-downstream-adoption
goal-certify
goal-runtime-final
```

这些 gate 必须支持：

```text
--goal-id GOAL-20260603-XLIB-GOALCLI-001
--mode FULL
--json
--write-evidence
```

写入证据时，目标路径为：

```text
release/evidence/goalcli/
```

## 5. v0.4.6 发布收敛条件

`v0.4.6` 发布必须满足：

```text
1. 源码、文档、schema、registry、Makefile 统一使用 goalcli。
2. Makefile 中机器 gate 委托给 goalcli。
3. command registry 和 Makefile baseline 覆盖新增 gate。
4. release metadata 升级到 v0.4.6。
5. PR 检查全部通过并合入 main。
6. GOWORK=off release preflight 通过。
7. tag v0.4.6 推送到 origin。
```

## 6. 禁止事项

```text
1. 不新增第二套 CLI。
2. 不把生成的 manifest 或 checksum 提交为源码。
3. 不把未验证的团队运行态当作发布证据。
4. 不绕过 Harness gate 直接声明完成。
5. 不在文档或代码中保留历史执行入口名称。
```

## 7. 验证证据

完成前至少需要以下 fresh evidence：

```text
go test ./...
go vet ./...
make docs-check contracts cli-contract command-registry makefile-baseline score
make release-check
make release-preflight VERSION=v0.4.6
gh pr checks <PR>
git ls-remote --tags origin refs/tags/v0.4.6
```

若某项验证因外部环境不可用无法执行，必须记录阻塞原因并使用最接近的本地 gate
作为补充证据；不能把未执行的命令写成已通过。
