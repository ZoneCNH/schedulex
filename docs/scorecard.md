# Release Scorecard

`go run ./cmd/goalcli score --min 9.8` 是发布质量分的可执行入口，默认输出 JSON，并在分数低于 `--min` 时以非零状态失败。分数不会替代 `make ci`、`make security` 或 `make release-final-check`，而是把 release manifest、workflow artifact、security gate、retrospective patch 和文档约束汇总为一个可审计信号。

## 语义边界

当前 `score` 是发布治理完整性分，不是运行时语义质量分。它主要证明必需文档、脚本、模板、manifest schema 和 gate wiring 存在并互相对齐；它不能替代单元测试、race、漏洞扫描、secret scan、integration 或 release Evidence 校验，也不能证明 public API 行为、覆盖率、benchmark 或下游真实运行质量。

因此，`score >= 9.8` 只能作为 release governance 信号使用。发布结论必须同时读取 `make ci`、`make security`、`make release-final-check`、测试覆盖证据和必要的人工审查结果。

Context Runtime v4.0 迁移期间，score 也不能替代 `context-standard`、`context-full` 或 `context-release` profile 结果。发布结论必须同时读取 `governance_runtime` manifest 字段、profile wrapper 和 registry bridge 的实际 gate 输出；物理 `.agent/context/*` packs/templates 未落地时，只能说明这些 packs/templates 尚未交付，不能用 scorecard 宣称它们已存在。

## 评分规则

当前满分为 10.0，默认发布阈值为 9.8。每个维度权重为 1.0，全部通过时得到 10.0；缺失任一维度会按权重扣分。评分维度如下：

| 维度 | 证明内容 |
| --- | --- |
| `scorecard_doc` | 本文件记录评分规则和阈值。 |
| `manifest_score_schema` | `release/manifest/template.json` 包含 `score`、`workflow_run_id`、`artifact_url`。 |
| `score_cli` | `cmd/goalcli` 提供 `score --min` 命令。 |
| `score_gate` | `Makefile` 在 release gate 中执行 score threshold；该维度验证 wiring，不验证业务语义。 |
| `manifest_min_score_verify` | `scripts/check_release_evidence.sh` 把 `RELEASE_EVIDENCE_MIN_SCORE` 传入 manifest 校验。 |
| `security_gate` | secret scan 脚本覆盖 provider token 与 private key 模式；真实安全结论仍以 `make security` 执行为准。 |
| `release_docs` | 发布文档要求 score、workflow run 和 artifact URL。 |
| `supply_chain_docs` | 供应链文档说明 score/workflow evidence。 |
| `retrospective_template` | retrospective 模板记录 Gate、Score 与 Patch rationale。 |
| `release_template` | release 模板要求本地 score 与 CI artifact evidence。 |

## Gate 契约

- `GOWORK=off make release-check` 会运行 `score-check`，默认要求 `score >= 9.8`。
- `GOWORK=off make release-final-check` 会再次运行 `go run ./cmd/goalcli score --min 9.8`，并要求 release manifest 内记录的 `score.value` 满足 `RELEASE_EVIDENCE_MIN_SCORE=9.8`。
- `release/manifest/latest.json` 会记录 `score` 和 `workflow`，其中 `workflow_run_id`、`artifact_name`、`artifact_url` 用于连接 CI artifact；本地运行时使用 `local:*` evidence URL。
- Standard Impact v4.0 必须区分 `context_runtime`、`governance_registry`、`repository_rules` 和 `downstream_context` 影响面，并把 `downstream_release_decision` 与 `repository_rules_release_decision` 写入 Evidence；score 不能替代这些 release decision。

## JSON 形状

```json
{
  "value": 10,
  "threshold": 9.8,
  "status": "passed",
  "dimensions": [
    {
      "name": "scorecard_doc",
      "weight": 1,
      "passed": true,
      "detail": "scorecard rubric is documented"
    }
  ]
}
```
