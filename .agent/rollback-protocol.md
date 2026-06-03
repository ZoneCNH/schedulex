# 回滚协议

1. 识别失败的 REQ、file set、command 和 owner。
2. 只回滚最小范围变更；如果 forward fix 更安全，则使用 forward fix。
3. 重新运行失败的 proof command 以及 `git diff --check`。
4. 在 `.agent/decision-log.md` 或 task result 中记录回滚决策。
5. 未经 leader 批准，不得回滚其他 worker 的 code、Makefile、CI、manifest 或 gate implementation。
