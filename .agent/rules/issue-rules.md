# Issue 规则

> 源自 Goal 完整规则 v1.0 §13

## RULE-ISSUE-001：Issue 必须从 Task 生成

Issue 必须包含：

```text
Goal ID
Task ID
Requirement ID
Acceptance Criteria
Implementation Scope
Files to Change
Commands to Run
Evidence Required
DoD
Risk
Rollback Plan
```

## RULE-ISSUE-002：Issue 必须有标准 Label

推荐 Label：

```text
goal          spec          design
task          harness       evidence
self-improving  release     risk
blocked       needs-research  needs-decision
```

## RULE-ISSUE-003：Issue 关闭必须有 Evidence

没有 Evidence 的 Issue 不允许关闭。
