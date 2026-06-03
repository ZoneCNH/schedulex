#!/usr/bin/env bash
set -euo pipefail

# RULE-WORKTREE-001: 禁止在 main/master 上直接开发

BRANCH="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "")"

if [ "$BRANCH" = "main" ] || [ "$BRANCH" = "master" ]; then
  echo "ERROR: 禁止在 $BRANCH 分支上直接开发。"
  echo "请使用 git worktree 创建独立开发分支。"
  echo "参考: .agent/rules/worktree-rules.md RULE-WORKTREE-001"
  exit 1
fi

echo "worktree-check: 当前分支 $BRANCH，通过。"
