# Scope Lock

`xlib-standard` 是标准、模板、generator、Harness 与 Evidence 的权威源。P0/P1/P2 gate 只能修改当前 worker worktree 中的源文件，不得写 `/home/xlib-standard` 主工作区。

强制范围：

- 禁止 x.go imports 进入基础库或 runtime 实现。
- 禁止读取、提交或打印真实 secrets；`/home/k8s/secrets/env/*` 只能作为路径名出现在文档中。
- `release/manifest/latest.json` 与 `.sha256` 是生成 Evidence，不进入源码。
