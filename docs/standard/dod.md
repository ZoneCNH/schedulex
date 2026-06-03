# 完成定义

DoD 按 Task、Issue、Goal 和 Release 四个层级递进。上层完成必须包含下层完成证据。

## Task DoD

- 变更范围明确，未混入无关重构。
- 相关测试或静态检查已运行。
- 文档或模板变更有对应入口。
- 已记录 known gap，或明确没有 known gap。

## Issue DoD

- 验收标准全部满足。
- Required gate 中与变更相关的命令通过。
- Review checklist 已完成。
- 无边界违规、密钥泄露或未解释的 flaky failure。

## Goal DoD

- ADR、PRD 或 Goal 文档中的所有必需项有实现或明确不适用理由。
- `GOWORK=off make ci`、`make integration`、`make evidence` 和 `make release-evidence-check` 有新鲜结果。
- 生成库或 downstream compatibility 的代表性验证已执行。
- 完成声明包含 `DONE with evidence:`。

## Release DoD

- `release/manifest/latest.json` 已生成并可校验。
- `GOWORK=off make release-check` 通过。
- `GOWORK=off go run ./cmd/goalcli score --min 9.8` 通过，manifest 内的 `score` 字段满足 release threshold。
- manifest 内的 `workflow_run_id`、`artifact_name`、`artifact_url` 已记录，能对齐 CI artifact 或明确的本地 Evidence URL。
- 发布前 `GOWORK=off make release-final-check` 通过。
- `GOWORK=off make release-preflight VERSION=<version>` 通过。
- CI artifact 保存 Evidence，不把 `latest.json` 提交进源码历史。

## 不能完成时

不得把 blocked 状态描述成完成。必须说明：

- 阻塞命令。
- 失败原因或缺失工具。
- 已完成的替代验证。
- 恢复所需的最小下一步。
