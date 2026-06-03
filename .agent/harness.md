# Harness 协议

## Full Goal Runtime v3.1 Gate

- Context Gate：确认当前目标、scope、worker ownership 和旧名迁移约束。
- Goal Gate：`GOAL-20260602-001` 与 `docs/goal.md` v2.9.3 Complete 的 REQ-001..REQ-010 对齐，并核对 `docs/project-analysis-20260602.md` 的 P0/P1/P2 与 52 项问题基线。
- Spec Gate：标准、模板、generator、Harness、Evidence 五类职责明确。
- Design Gate：模块边界、下游矩阵、x.go 集成边界和 secret policy 明确。
- Plan Gate：任务切片不越权；module/name、gate CLI、manifest/score 等由对应 worker 负责。
- Task Gate：每个任务有 traceability 和验证命令。
- Implementation Gate：最小变更、无旧主身份回归、无 secret 内容。
- Test Gate：docs、中文文档规则、Go、boundary、contract、release 和 downstream gate 有新鲜证据。
- Evidence Gate：manifest/checksum/score/日志满足 Evidence Protocol。
- Review Gate：review template 覆盖边界、旧名、secret、score、kernel downstream。
- Release Gate：`release-final-check`、`release-preflight` 和 `goalcli score --min 9.8`。
- Retrospective Gate：失败或风险必须写入 retrospective、prompt/harness/rule patch。
