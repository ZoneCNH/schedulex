# schedulex 标准索引

[`github.com/ZoneCNH/schedulex`](https://github.com/ZoneCNH/schedulex) 当前交付目标是 `pkg/schedulex` 的 L1 deterministic scheduler v0.1.0。本索引用于定位 L1 调度器的 API、边界、gate 和 Evidence 规则。

## 必读标准

- [docs/standard/schedulex.md](schedulex.md)：L1 scheduler 的公共 API、确定性时间、trigger、misfire、overlap、lock interface、事件和 gate。
- [模块边界](module-boundary.md)：基础库不得依赖 `x.go`、L2 runtime、真实密钥路径或应用组合层。
- [下游同步策略](../downstream-sync-policy.md)：标准变更到 L1/L2 基础库和 `x.go` 消费方的同步规则。
- [Release 标准](release-standard.md)：release manifest、checksum、preflight 和 final check 规则。
- [Evidence 协议](evidence-protocol.md)：`release/manifest/latest.json`、`release/manifest/latest.json.sha256` 和 DONE 声明。

## Gate

发布式验证必须运行：

```bash
GOWORK=off make docs-check
GOWORK=off make release-final-check
GOWORK=off make release-preflight VERSION=v0.1.0
```

完整 release Evidence 还需要 `release/manifest/latest.json`、`release/manifest/latest.json.sha256`、public API snapshot、trigger/misfire/timezone-DST golden、downstream smoke 和最终 `DONE with evidence:` 声明。
