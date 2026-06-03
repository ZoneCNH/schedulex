# Harness 规则

> 源自 Goal 完整规则 v1.0 §11

## RULE-HARNESS-001：Harness Gate 是阶段裁判

每个阶段必须有 Gate：

```text
Context Gate     Goal Gate       Spec Gate
Design Gate      Plan Gate       Task Gate
Issue Gate       Worktree Gate   Commit Gate
PR Gate          CI Gate         Evidence Gate
Review Gate      Release Gate    Retrospective Gate
Self-improving Gate
```

## RULE-HARNESS-002：Gate 必须有标准结构

```yaml
gate_id:
name:
type: semantic | executable | hybrid
severity: P0 | P1 | P2 | P3
target:
rules:
commands:
pass_condition:
fail_condition:
evidence_required:
```

## RULE-HARNESS-003：P0 Gate 失败必须阻断

P0 Gate 失败时：

```text
禁止 commit
禁止 PR ready
禁止 merge
禁止 release
禁止 DONE
```

## RULE-HARNESS-004：Harness 结果必须归档

必须生成：

```text
reports/context-check.json
reports/spec-check.json
reports/design-check.json
reports/task-check.json
reports/worktree-check.txt
reports/evidence-check.json
reports/release-check.json
```
