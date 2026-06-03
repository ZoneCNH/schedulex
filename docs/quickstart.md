# 快速开始

本指南用于在本地复核 xlib-standard 的标准、门禁与证据链，不代表下游 kernel/configx 已通过。

## 环境

- Go 版本遵循 `.tool-versions` 与 CI 配置：`1.23.x`。
- 治理、发布和证据命令必须显式使用 `GOWORK=off`。

## 常用命令

```sh
GOWORK=off make docs-check
GOWORK=off make governance-check
GOWORK=off make p1-governance-check
GOWORK=off make p2-runtime-check
GOWORK=off make release-check
```

## 发布证据

`release/manifest/latest.json` 与校验和是运行时生成产物，按 `.gitignore` 不提交。需要复核时重新运行对应 release/evidence 命令，并保留 CI artifact 或本地输出作为证据。

## 下游采纳

下游 kernel/configx 的通过状态必须来自真实下游运行结果；本仓库中的占位目录和文档只说明结构契约，不可作为通过证明。
