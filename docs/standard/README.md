# xlib-standard 基础库标准索引

[`xlib-standard`](https://github.com/ZoneCNH/xlib-standard) 是基础库 Standard Source、Go Reference Template、Generator、Harness 和 Evidence Runtime 的统一仓库。旧 `baselib-template` 名称只作为迁移兼容上下文保留；旧 `foundationx` 默认下游名迁移为 `kernel`。

## 必读标准

- [基础库总标准](xlib-standard.md)：公共 API、配置、错误、健康检查、metrics、测试、安全和发布规则。
- [仓库角色](repository-roles.md)：`xlib-standard`、`kernel`、各生成库和 `x.go` 的职责。
- [分层](layering.md)：Standard、L0、L1、L2、应用组合层关系。
- [模块边界](module-boundary.md)：允许/禁止内容、module path 和 `x.go` 边界。
- [完成定义](dod.md)：基础库 DONE with evidence 的最低标准。
- [Harness gate](harness-gates.md)：required、extended、generator、docs、score 和 final gate。
- [Evidence 协议](evidence-protocol.md)：release/manifest/template.json、release/manifest/latest.json、artifact_url、workflow_run_id、sha256 和 DONE 声明。
- [Release 标准](release-standard.md)：release/manifest/latest.json.sha256、preflight 和 final check。
- [安全与密钥](security-and-secret-policy.md)：禁止泄露生产密钥和 `/home/k8s/secrets/env/*` 内容。
- [模板生成契约](template-generation-contract.md)：module path、package name、README/docs 替换规则。
- [下游兼容性](downstream-compatibility.md)：生成库兼容窗口和变更级别。

## 当前 v3.1 补充入口

- [Goal Runtime Canonical 标准](../../.agent/standard/goal-runtime-canonical.md)：Goal Runtime 唯一权威规格（8 条铁律 + 9 层架构 + v0.1.0 五主线），原始演进合集见 `.agent/inbox/`。
- [goalcli 命名合约](../../.agent/standard/goalcli-mapping.md)：单一命名合约，`goalcli` 是标准合约、机器执行面和本仓库实现入口。
- [下游矩阵](../downstream-matrix.md)：`kernel` 与目标基础库的 module/package/layer/dependency 矩阵。
- [下游同步策略](../downstream-sync-policy.md)：标准变更到 `kernel`、L1/L2 基础库和 `x.go` 的同步规则。
- [x.go 集成边界](../xgo-integration-boundary.md)：调用方密钥路径和组合边界。
- [测试策略](../testing.md)：单元、示例 smoke、release quality 和 release manifest fixture 隔离要求。
- [供应链与 Evidence](../supply-chain.md)：workflow Action SHA pinning、固定 `govulncheck` 版本、manifest 校验和 CI artifact 对齐。
- [Release Scorecard](../scorecard.md)：`goalcli score --min 9.8` 的评分维度、阈值和语义边界。
- [独立审计 2026-06-02](../independent-audit-20260602.md)：审计发现、修复状态和剩余远端验证缺口。
- [迁移指南](../migration/baselib-template-to-xlib-standard.md)：旧名迁移规则。
- [目标 ADR-001](../adr/ADR-20260602-001-xlib-standard-role.md)：合并五类职责的身份决策。
- [目标 ADR-002](../adr/ADR-20260602-002-kernel-rename.md)：默认下游名迁移到 `kernel`。

## Gate

发布式验证必须至少运行：

```bash
GOWORK=off make dependency-check
GOWORK=off make standard-impact-check
GOWORK=off make docs-check
GOWORK=off go run ./cmd/goalcli score --min 9.8
GOWORK=off make release-check
```

完整 release Evidence 还需要 `release/manifest/latest.json`、`release/manifest/latest.json.sha256`、`release/standard-impact/latest.md`、`downstream_sync_required` 结论、manifest 内的 `score` 与 `workflow` 字段、CI artifact 和 `DONE with evidence:` 声明。Fuzz smoke 默认使用 `FUZZ_SMOKE_TIME=10s`，加长时必须记录到 Evidence。

CI、Release Check 和 Security workflow 的第三方 Action 必须 pin 到 40 位 commit SHA，并保留来源 tag 注释。`govulncheck` 发布门禁基线为 `golang.org/x/vuln/cmd/govulncheck@v1.3.0`；release manifest 测试必须在临时 fixture 仓库构造 `.omc` state，不得读取当前工作区的 Agent 运行态。
