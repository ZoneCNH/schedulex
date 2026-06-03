# PR 规则

> 源自 Goal 完整规则 v1.0 §15

## RULE-PR-001：PR 必须是可审查交付单元

PR 必须包含：

```text
Goal
Related Issues
Requirements Covered
Changes
Tests
Evidence
Risk
Rollback
Checklist
```

## RULE-PR-002：PR 必须包含 Traceability

```text
Requirement | AC | Task | Test | Evidence | Status
```

## RULE-PR-003：PR 合并条件

必须满足：

```text
CI passed
Harness passed
Evidence complete
Traceability complete
No P0 risk open
Review approved
main up to date
```

## RULE-PR-004：禁止直接合并未验证 PR

没有 Evidence / Harness / Review 的 PR 不允许 merge。
