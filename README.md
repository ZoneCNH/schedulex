# xlib-standard

`xlib-standard` 是基础库标准与交付运行时仓库，承担五类职责：**Standard Source**、**Go Reference Template**、**Generator**、**Harness** 和 **Evidence Runtime**。它把基础库的公共 API、配置、错误、健康检查、metrics、测试、release Evidence、Goal Runtime 和下游生成规则放在同一套可验证工件中维护。

旧名 `baselib-template` 和示例名 `foundationx` 只允许出现在迁移文档语境中；新的默认下游集成目标是 `kernel`，生成库包括 `configx`、`observex`、`testkitx`、`postgresx`、`redisx`、`kafkax`、`taosx`、`ossx` 和 `clickhousex`。

标准源仓库 URL 为 [`xlib-standard`](https://github.com/ZoneCNH/xlib-standard)。本仓库不再把标准源与模板实现拆成两个角色：标准文本、模板、generator、Harness gate 和 Evidence runtime 必须一起通过 release gate 验证。

## 五类职责

- **Standard Source**：维护基础库 P0 标准、仓库角色、分层、模块边界、DoD、安全、release 和 Evidence 协议。
- **Go Reference Template**：提供可编译参考包 `pkg/templatex`、内部辅助、examples、contracts 和 testkit，用于证明标准可落地。
- **Generator**：通过 [docs/generation.md](docs/generation.md) 与 `scripts/render_template.sh` 渲染具体基础库 module path、package name、README、docs 和 contracts。
- **Harness**：通过 Makefile、scripts、CI 和 [.agent/harness.yaml](.agent/harness.yaml) 固化 required、extended、docs、boundary、integration、score 和 final gate。
- **Evidence Runtime**：通过 [docs/standard/evidence-protocol.md](docs/standard/evidence-protocol.md)、[docs/release.md](docs/release.md)、[.agent](.agent/) 和 `release/manifest/latest.json` 记录可追溯完成状态。

## 非目标

- 不依赖 `x.go`，也不把 `x.go` 作为基础库构建前提。
- 不包含 `x.go` 业务模型、业务 repository、业务消息 schema 或应用 wiring。
- 不隐式读取生产密钥，不把 `/home/k8s/secrets/env/*` 的内容写入源码、README、测试日志、manifest 或 PR 描述。
- 不创建隐藏全局客户端、不可关闭后台进程或真实基础设施 runtime。
- 不把旧 `baselib-template` / `foundationx` 叙事继续作为主身份。

## 标准结构

- `pkg/templatex`：公共包 API 的可编译参考实现；渲染后会移动到 `pkg/<package-name>`。
- `internal/`：脱敏、校验和运行时说明等内部辅助代码。
- `testkit/`：可复用测试夹具和断言。
- `examples/`：最小使用示例。
- `contracts/`：JSON schema 和 metrics contract。
- `docs/`：规格、设计、API、配置、测试、标准、迁移和发布文档。
- `scripts/`：Harness gate 与 Evidence 脚本。
- `.agent/`：Full Goal Runtime v3.1 工件、Evidence、评审、发布、回滚和复盘模板。
- `release/manifest/`：release manifest 模板；`latest.json` 由 release gate 生成并作为 Evidence artifact 保存。

## 文档入口

- [基础库标准索引](docs/standard/README.md)：P0 标准入口，覆盖仓库角色、分层、DoD、Harness、Evidence、release、安全和 generator 契约。
- [基础库总标准](docs/standard/xlib-standard.md)：同步 [`xlib-standard`](https://github.com/ZoneCNH/xlib-standard) 的公共 API、配置、错误、健康检查、metrics、测试、安全和发布规则。
- [仓库角色](docs/standard/repository-roles.md)：区分 `xlib-standard`、`kernel`、生成基础库和 `x.go`。
- [模块边界](docs/standard/module-boundary.md)：定义标准、模板、generator、Harness、Evidence 与下游库边界。
- [下游矩阵](docs/downstream-matrix.md)：列出 `kernel` 与所有目标库的 module path、package、layer、允许依赖和禁止依赖。
- [下游同步策略](docs/downstream-sync-policy.md)：定义 `xlib-standard` 变更如何同步到 `kernel`、L1/L2 基础库，以及 `x.go` 的消费方边界。
- [x.go 集成边界](docs/xgo-integration-boundary.md)：说明 `x.go` 只能作为调用方组合层，基础库不得反向依赖。
- [迁移指南](docs/migration/baselib-template-to-xlib-standard.md)：记录旧名到新身份的迁移规则。
- [Harness gate](docs/standard/harness-gates.md)：required、extended、generator、docs、score 和 final gate 命令。
- [Evidence 协议](docs/standard/evidence-protocol.md)：`DONE with evidence:` 和 release manifest 要求。
- [测试策略](docs/testing.md)：单元、示例 smoke、release quality 和 release manifest fixture 隔离要求。
- [安全与密钥策略](docs/standard/security-and-secret-policy.md)：`govulncheck`、secret scan 和 Agent runtime 目录排除边界。
- [供应链与 Evidence](docs/supply-chain.md)：workflow Action SHA pinning、固定 `govulncheck` 版本、release manifest 和 CI artifact 对齐。
- [Release Scorecard](docs/scorecard.md)：`goalcli score --min 9.8` 的评分维度、阈值和语义边界。
- [发布](docs/release.md)：`release-check`、manifest 字段和 Evidence 规则。
- [独立审计 2026-06-02](docs/independent-audit-20260602.md)：独立审计发现、修复状态和剩余验证缺口。
- [项目分析快照 2026-06-02](docs/project-analysis-20260602.md)：`v0.3.7` 发布/分析快照；当前治理主基线仍以 [目标文档](docs/goal/goal.md) v2.9.3 Complete 和 [.agent/traceability-matrix.md](.agent/traceability-matrix.md) 为准。
- [结构性问题清单 2026-06-02](docs/structural-issues-20260602.md)：记录架构、治理和交付风险的结构化问题清单。
- [.agent 真相状态文件](.agent/truth-state.yaml)：汇总当前治理、命令实现、release gate、Evidence 可用性和下游采纳状态口径。

## 命令

本地运行完整 gate 前需要安装 `golangci-lint` 和 `govulncheck`；CI 会显式安装这两个工具。缺少任一工具时，`make lint` 或 `make security` 必须失败，不允许把必需 gate 记录为跳过。

### 首次 clone 必跑

新协作者 clone 仓库后必须立即执行：

```bash
make install-hooks   # 启用 .githooks 本地 P0 防线（RULE-WORKTREE-001 + RULE-SECRET-001）
make doctor-hooks    # 验证 core.hooksPath=.githooks 已生效
make sync-main       # 拉取并 fast-forward 本地 main（RULE-MAIN-SYNC-002）
```

`make install-hooks` 把 `git config core.hooksPath` 指向仓库内的 `.githooks/` 目录。**未启用 hooks 时，`pre-commit` 与 `pre-push` 不会被 Git 调用，本地 P0 防线（禁止在 main commit、secret 提前拦截）形同虚设。** 此外，`go run ./cmd/goalcli doctor` 会在 details 中报告当前 hooks 启用状态，配合 `make doctor-hooks` 形成自检闭环。`make ci` 链首位的 `doctor-hooks-local` 也会在本地环境强制 fail-fast（CI 环境通过 `$CI` / `$GITHUB_ACTIONS` 自动跳过）。

### 标准 gate

```bash
make ci
make ci-extended
GOWORK=off make dependency-check
GOWORK=off make standard-impact-check
GOWORK=off make docs-check
XLIB_CONTEXT=release_verify GOWORK=off make release-check
XLIB_CONTEXT=release_verify GOWORK=off make release-final-check
XLIB_CONTEXT=release_verify GOWORK=off make release-preflight VERSION=v0.4.6
make evidence
```

`release-check` 和 `release-check-extended` 已依赖 `dependency-check`、`standard-impact-check` 和 `docs-check`，用于在生成 Evidence 前确认依赖漂移自动化、标准影响报告、标准文档入口、下游同步策略、链接、模板占位符、当前命名、关键文本和 release manifest 协议没有漂移。`dependency-check` 读取 `renovate.json`、`.github/dependabot.yml` 和 `go.mod`；`standard-impact-check` 生成 `release/standard-impact/latest.md`，并把 `downstream_sync_required`、`downstream_release_decision`（只允许 `required` / `not_required`）和 `repository_rules_release_decision`（只允许 `audit_required` / `not_required`）结论交给 release manifest。`docs-check` 是结构性 gate，不替代人工语义审查。

Release gate 还必须执行 `GOWORK=off go run ./cmd/goalcli score --min 9.8`。GitHub Actions workflow 引用的第三方 Action 必须固定为 40 位 commit SHA 并保留来源 tag 注释；CI、Release Check 和 Security workflow 安装 `govulncheck` 时使用固定基线 `golang.org/x/vuln/cmd/govulncheck@v1.3.0`，不得用 `@latest` 作为发布门禁配置。

生成 `kernel` 示例：

```bash
scripts/render_template.sh \
  --module-name kernel \
  --module-path github.com/ZoneCNH/kernel \
  --package-name kernel \
  --out ../kernel
```

发布式验证必须使用 `GOWORK=off`，避免父级或本地 `go.work` 改写 module 解析并掩盖模板独立性问题：

```bash
GOWORK=off make docs-check
XLIB_CONTEXT=release_verify GOWORK=off make release-check
```

## Evidence

完成需要 release manifest 和 CI Evidence。`release/manifest/latest.json` 是生成产物，不提交到源码历史；对应的 `release/manifest/latest.json.sha256` 也是生成产物，两者都必须保持在 `.gitignore` 中。manifest 会记录 module、commit、tree SHA、源码摘要、contract 指纹、`dependencies`、`tools`、生成时间、工作区状态、gate 结果、`standard_impact`、`downstream_sync_required`、`generator_evidence`、`score`、`workflow` 和这两个 Evidence artifact；其中 `standard_impact.downstream_release_decision` 只能使用 `required` 或 `not_required`，`standard_impact.repository_rules_release_decision` 只能使用 `audit_required` 或 `not_required`。`release-check` 会生成并校验 checksum，CI 会上传两者作为 artifact。`make release-evidence-check` 会验证 manifest 与当前仓库事实一致，`make release-final-check` 会额外要求工作区为 `clean`。Release manifest 测试必须在临时 fixture 仓库中构造所需 `.omc` state，不得依赖当前工作区的 Agent 运行态文件。最终完成声明必须包含 `DONE with evidence:`。

Full Goal Runtime v3.1 位于 [.agent](.agent/)，其中 [goal-runtime](.agent/goal-runtime.md)、[object-model](.agent/object-model.md)、[state-machine](.agent/state-machine.md)、[traceability-matrix](.agent/traceability-matrix.md)、[harness](.agent/harness.yaml)、[evidence-protocol](.agent/evidence-protocol.md)、[release-template](.agent/release-template.md)、[retrospective-template](.agent/retrospective-template.md)、[risk-register](.agent/risk-register.md)、[decision-log](.agent/decision-log.md)、[rollback-protocol](.agent/rollback-protocol.md) 和 patch 文档用于把标准、执行、评审、发布和复盘连接到同一套 Evidence 协议。

## Smoke 覆盖

`go test ./...` 必须覆盖公共包、`internal/`、`contracts/`、`testkit/` 和 `examples/`。当前示例 smoke 测试会验证 `examples/basic` 输出模块名、`examples/config` 输出脱敏值、`examples/health` 输出健康状态，防止文档示例和模板行为漂移。

`scripts/run_fuzz_smoke.sh` 默认执行快速 fuzz smoke，`FUZZ_SMOKE_TIME` 未设置时每个 fuzz target 使用 `10s`。需要深度 fuzz 时显式设置更长时间，例如 `FUZZ_SMOKE_TIME=2m make fuzz-smoke`，并在最终 Evidence/DONE 说明中记录该时间配置。
