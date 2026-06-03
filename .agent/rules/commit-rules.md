# Commit 规则

> 源自 Goal 完整规则 v1.0 §14

## RULE-COMMIT-001：Commit 必须绑定 Task / Issue / Evidence

推荐格式：

```text
<type>(<scope>): <summary>

Refs: TASK-xxx
Closes: #123
Evidence: EVID-xxx
```

## RULE-COMMIT-002：禁止无语义提交

禁止：

```text
fix
update
wip
stuff
misc
temp
```

## RULE-COMMIT-003：禁止大杂烩提交

一个 commit 应该对应一个明确变更意图。

## RULE-COMMIT-004：Commit 前必须通过本地 Gate

至少：

```text
make worktree-check
make lint
make test
make evidence-check
```
