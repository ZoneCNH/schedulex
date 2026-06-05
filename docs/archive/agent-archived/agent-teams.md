# Agent Teams

当任务横跨标准文档、Harness、generator、release 和治理模板时，使用 team 模式拆分检查面。Leader 负责整合、编辑和最终验证。

## 建议分工

| Lane | 关注点 | 产出 |
| --- | --- | --- |
| Standard docs | `docs/standard/`、README 导航、仓库角色 | 缺口清单和文档补齐建议 |
| Harness/Evidence | Makefile、scripts、manifest、release gate | gate 覆盖和 Evidence 风险 |
| Governance | `.agent/`、Issue/PR 模板、review/retrospective | 协作模板和完成协议 |

## 运行规则

- worker 只报告事实、风险和建议，不直接改写全局目标。
- leader 保持一个最终计划，避免多个 worker 同时改同一文件。
- 所有结果必须回到 required gate 和 `DONE with evidence:`。
- team 结束前检查 pending、in_progress、failed 任务数量。

## 最小 Evidence

- team status 或 worker 报告显示无未处理任务。
- leader 运行本地验证命令。
- final response 汇总 changed files、验证结果和 known gaps。
