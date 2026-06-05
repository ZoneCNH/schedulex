# Agent Team Contract（xlib-standard v2.9.3）

本文件记录 P1-001 Agent Team Contract 的本地可验证契约，用于 `goalcli agent-team-contract` 与 team worker 验收。

## 角色与职责

- Lead：维护目标、任务拆分、最终审查与验收结论；禁止覆盖用户拥有的主工作区修改。
- Executor Worker：仅在分配的 worker worktree 内实现、测试、提交并报告证据。
- Verifier Worker：复核命令输出、证据链、回归风险与剩余缺口；不得自审自己的实现。

## 范围锁定

- 所有实现必须使用 team 自动创建的 worker worktree，不写入 `/home/xlib-standard` 主工作区。
- 共享文件修改必须来自已分配任务；发生冲突时先上报 leader。
- 禁止读取真实 secrets，禁止新增依赖，禁止 `x.go` imports，`release/manifest/latest.json` 仅作为生成物。

## Gate 与证据

- 必须在 `GOWORK=off` 下执行相关 goalcli、Makefile、测试与治理命令。
- 完成报告必须包含 changed files、验证命令、PASS/FAIL 结果、剩余风险。
- Team lifecycle 以 OMX task claim/status 为准，完成前必须提交 worker worktree 变更。
