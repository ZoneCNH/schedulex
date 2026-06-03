# Agent Team Contract

Goal v2.9.3 team 由 leader 负责计划、集成和最终 Review；worker 只执行被分配的 slice，并通过 Evidence 报告结果。worker 不重写全局计划，不覆盖主工作区，不读真实 secrets。

每个 worker 完成时必须提供：

- changed files；
- 验证命令和 PASS/FAIL 输出摘要；
- 剩余风险；
- `DONE with evidence:` 形式的 Evidence 链接或本地命令证据。

leader 在合并前检查 `.agent/issue-registry.yaml`、`.agent/command-registry.yaml`、Makefile gates、CI 和 contracts 是否一致。
