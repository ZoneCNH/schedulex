# goalcli v0.1.0 Agent Team Context

生成时间：2026-06-03T00:53:02Z
更新日期：2026-06-04
Workspace：`/home/xlib-standard`
Goal：使用 agent teams 执行 `.worktree/goalcli-v0.1.0-plan.md`

## 目标结果

通过团队分工把 `goalcli v0.1.0` 的最小可验收面落到 `xlib-standard`：

```text
1. 收敛执行边界，避免多个执行入口并列。
2. 让 Makefile、registry、contract、tests 和 docs 指向 goalcli。
3. 为每个完成声明生成 fresh evidence。
4. 以 v0.4.6 发布验证和 tag 推送作为停止条件。
```

## 当前事实

```text
1. `cmd/goalcli` 是唯一机器执行面。
2. `docs/standard/goalcli-runtime.md` 和
   `docs/standard/goalcli-cli-contract.md` 已进入标准路径。
3. `docs/adr/ADR-20260603-001-goalcli-runtime.md` 记录运行时决策。
4. `.agent/harness.yaml`、command registry 和 Makefile baseline 需要与
   `goalcli` gate 保持一致。
5. release evidence 写入 `release/evidence/goalcli/`。
6. `.omx/` 与 `.worktree/` 文件默认被忽略；被源码 gate 引用的 source-only
   文件必须显式入库。
```

## 团队执行分工

Lane A - 文档与权威路径：

```text
1. 对齐 plan、ADR、standard、CLI contract 和 release evidence。
2. 文档默认使用中文，保留代码标识符、命令、路径和协议固定短语。
3. 清理历史执行入口名称。
```

Lane B - 运行时与命令实现：

```text
1. `cmd/goalcli` 承接 Makefile gate。
2. command registry 覆盖新增命令。
3. Makefile baseline 反映真实委托关系。
```

Lane C - 测试与证据：

```text
1. 更新 focused regression tests。
2. 运行 targeted Go tests、repo-wide tests 和 release gates。
3. PR 检查全绿后合入 main，再执行 release preflight。
```

## 硬约束

```text
1. 不提交生成产物 `release/manifest/latest.json` 或 checksum。
2. 不复用 inactive team state 作为当前发布证据。
3. 不声明未执行的 gate 已通过。
4. 不绕过 GitHub required checks 直接发布 main。
5. 保留与本任务相容的用户或 team 工作树改动。
```

## 停止条件

```text
1. PR 检查全部通过并合入 main。
2. `GOWORK=off VERSION=v0.4.6 make release-preflight` 通过。
3. `v0.4.6` tag 存在于 origin。
4. 最终报告包含 DONE with evidence: 和命令证据。
```
