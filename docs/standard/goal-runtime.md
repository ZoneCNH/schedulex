# Goal Runtime

Goal runtime 以 `cmd/goalcli` 为唯一 Go runtime 入口，Makefile target 是人机友好包装。`runtime-health` 和 `goal-runtime` gate 验证以下事实：

- runtime 命令表来自 `.agent/command-registry.yaml`；
- P0/P1/P2 target 来自 `.agent/makefile-target-registry.yaml`；
- `release/manifest/latest.json` 是生成 Evidence；
- worker worktree 与 CI checkout 可以运行相同 dry-run gate；
- 不读取真实 secrets，不引入 x.go imports。

Runtime file ownership：`cmd/goalcli/main.go` 拥有 CLI dispatch；`.agent/*registry*.yaml` 拥有治理 SSOT；`docs/standard/goalcli-cli-contract.md` 拥有用户可读 contract。
