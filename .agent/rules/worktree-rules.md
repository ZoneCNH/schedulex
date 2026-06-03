# Worktree 规则

> 源自 Goal 完整规则 v1.0 §10

## RULE-WORKTREE-001：禁止 main 开发

main / master 只能作为同步基线与发布基线。禁止在 main / master 上直接开发、commit、push。

## RULE-WORKTREE-002：所有开发必须使用 git worktree

每个 Goal / Issue / Task 必须创建独立 worktree。

推荐结构：

```text
~/code/<repo>                               # main worktree，只同步
~/code/.worktrees/<repo>/<goal>/<task>      # 开发 worktree
```

## RULE-WORKTREE-003：worktree 分支必须绑定 Goal / Task

分支命名：

```text
goal/<GOAL-ID>/<TASK-ID>
issue/<ISSUE-ID>
task/<TASK-ID>
```

## RULE-WORKTREE-004：必须提供 worktree gate

必须有：

```text
make worktree-check
scripts/harness/no-main-dev.sh
.githooks/pre-commit
.githooks/pre-push
```

## RULE-WORKTREE-005：PR 合并后必须清理 worktree

```bash
git worktree remove <path>
git worktree prune
```
