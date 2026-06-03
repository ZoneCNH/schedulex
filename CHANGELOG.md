# 变更日志

## 未发布

暂无。

## v0.4.6 - 2026-06-04

### 治理

- 将 Goal Runtime、Harness 和文档中的执行入口统一为 `goalcli`，移除旧命名面的并行入口。
- 同步 `cmd/goalcli`、`internal/goalcli`、contract schema、release Evidence 和标准文档路径，确保 gate 与文档引用一致。
- 将项目发布版本、release manifest 默认版本、`goalcli` 治理版本和 release preflight 示例同步到 `v0.4.6`。

### 兼容性

- 本版本不改变 `pkg/templatex` 公共 API 形状；`Version` 元数据同步到 `v0.4.6`。

## v0.4.5 - 2026-06-03

### 修复

- 修正 README、downstream sync policy 和项目分析文档中的 `docs/goal.md` 迁移后链接，统一指向 `docs/goal/goal.md`。
- 移除 `goalcli` 测试中不再使用的 helper，保持测试文件无死代码。

### 治理

- 为 CI、Goal Gate、Integration、Release、Security 和 Worktree Guard workflow 增加 concurrency 控制。
- 将 GitHub Actions Go 安装统一为读取 `go.mod` 的 `go-version-file`，并启用 `setup-go` 内建 module cache。
- 将项目发布版本、release manifest 默认版本、goalcli 治理版本和 release preflight 示例同步到 `v0.4.5`。

### 兼容性

- 本版本不改变 `pkg/templatex` 公共 API，仅更新版本元数据、CI 治理和文档链接。

## v0.4.3 - 2026-06-03

### 修复

- 将 `downstream-debt` alias 收敛为 downstream 专用检查，避免误触发 architecture debt。
- 补齐 downstream integration 对 debt Evidence 与 checksum gate 的要求，并锁定渲染后债务证据产物不得进入下游源码。
- 将 downstream release/integration 覆盖目标同步为 `kernel`、`configx` 和 `redisx`。

### 治理

- 将项目发布版本、release manifest 默认版本、goalcli 治理版本和 release preflight 示例同步到 `v0.4.3`。

### 兼容性

- 本版本不改变 `pkg/templatex` 公共 API，仅收紧 downstream debt 和 release Evidence 治理。

## v0.4.2 - 2026-06-03

### 修复

- 渲染模板时保留 `release/manifest/template.json`，避免下游 release/version gate 缺少 manifest 模板。
- 渲染后占位符扫描豁免 release manifest 模板中的有意占位符，保持下游检查与发布模板语义一致。

### 治理

- 将项目发布版本、release manifest 默认版本、goalcli 治理版本和 release preflight 示例同步到 `v0.4.2`。

### 兼容性

- 本版本不改变 `pkg/templatex` 公共 API，仅修复 release Evidence/generator 交付口径。

## v0.4.1 - 2026-06-03

### 治理

- 对齐 release 版本口径到 `v0.4.1`，同步 `templatex.Version`、release manifest 默认版本、preflight 命令示例和 Harness 记录。

### 兼容性

- 本版本仅更新发布治理与版本元数据，不改变 `pkg/templatex` 公共 API 行为。

## v0.4.0 - 2026-06-03

### 修复

- Release manifest 校验现在会拒绝非法的 `standard_impact.downstream_release_decision` 和 `standard_impact.repository_rules_release_decision` 枚举值，避免发布 Evidence 接受漂移口径。
- `docs-check` 同步锁定 README、发布文档、Evidence protocol、downstream sync policy 和 Harness Gate 中的 release decision 允许值说明。

### 兼容性

- 本版本不改变 `pkg/templatex` 公共 API，仅收紧发布 Evidence 校验与文档门禁。

## v0.3.8 - 2026-06-02

### 治理

- 新增 MVA 真相态、命令实现、Evidence usability、release required gates 和 downstream adoption 状态文件。
- 新增 `docs/standard/truth-state.md`，记录首个 PR 的真相态语义、缺口和验收命令。
- 对齐 MVA 验收证据命令为 `GOWORK=off make governance-check`，避免与 release verify 上下文混用。

### 兼容性

- 本版本仅包含治理和文档更新，不改变 `pkg/templatex` 公共 API。

## v0.3.7 - 2026-06-02

### 治理

- 建立 `docs/goal.md` 治理目标基线，并新增 `docs/project-analysis-20260602.md` 记录深度检查结论。
- 对齐 `.agent` traceability matrix、Harness、`docs/spec.md` 与 downstream sync policy 的 P1/P2 gate 口径。
- 更新 `AGENTS.md` 并新增 `CLAUDE.md`，保持 Codex/Claude 协作入口一致。

### 兼容性

- 本版本仅包含治理和文档更新，不改变 `pkg/templatex` 公共 API。

## v0.3.6 - 2026-06-02

### 修复

- Release manifest 测试改用临时 git fixture 构造 `.omc/state/agent-replay-fixture.jsonl`，避免依赖本地 Agent 运行态文件。
- GitHub Actions workflow 固定 `checkout`、`setup-go`、`cache` 和 `upload-artifact` 的 40 位 commit SHA，并将 `govulncheck` 固定为 `v1.3.0`。
- Secret Gate 同时排除 `.omc` 和 `.omx` 本地运行态目录，避免扫描 Agent 状态文件时产生误报。
- `goalcli` dry-run verifier 在具备 manifest 覆盖时返回 `passed`，避免 `--verify` 模式继续报告 planned gap。
- `downstream-baseline`、`downstream-adoption` 和 `upgrade-standard` 默认使用 manifest-only dry-run，只有显式传入 `--repo` 时才检查本地 downstream 路径。
- Makefile baseline 将 `security` 目标对齐为 `$(GOALCLI) security`，并把 `execution-context` 纳入强制 target 覆盖。

### 治理

- P2 Runtime Dry Run 新增 `execution-context` gate，并同步 `.agent` registry、Harness 文档和 docs-check 漂移检测。

### 测试

- 补齐 `internal/releasequality` 对 `Compute`、`Verify` 和 `Marshal` 的单元测试。
- 补充 `goalcli` 对 `execution-context` baseline 缺口、manifest-backed verify 通过路径和 downstream 显式 repo gap 的回归测试。

### 文档

- 对齐 `docs/independent-audit-20260602.md` 的修复状态，并补充 score 语义边界、workflow pinning 和固定工具版本要求。
- 更新 Harness Gate 说明，明确 P2 Runtime Dry Run 覆盖 runtime-file-ownership 与 execution-context。

## v0.3.5 - 2026-06-02

### 新增

- 新增 `cmd/goalcli` gate 路由入口，统一封装 `release-final-check`、`release-evidence-check` 和 `score` 等发布前检查。
- 新增 release scorecard 文档与 `internal/releasequality` 评分实现，将 manifest、workflow artifact、安全门禁、复盘补丁和文档约束汇总为可执行分数。
- 新增 downstream compatibility matrix，记录 `kernel`、`foundationx` 和 `corekit` 渲染后的测试、contracts、boundary 与 Evidence 验证结果。

### 治理

- `release-final-check` 强制校验 scored release Evidence，避免仅凭局部 gate 结果推进发布。
- 发布文档、Agent 运行时文档和标准文档统一到 `xlib-standard` 命名与 release gate 口径。
- GitHub CI 与 release workflow 对齐 `GOWORK=off`、docs-check、security、contracts、boundary 和 release manifest 校验。

### 验证

- 发布前已运行 `GOWORK=off make release-final-check`。
- 发布前已运行 `go run ./cmd/goalcli score --min 9.8`，当前质量分为 10。
- 发布前已运行 `go run ./cmd/goalcli release-evidence-check`，确认 `release/manifest/latest.json` 通过校验。

## v0.3.0 - 2026-06-01

### 新增

- Release Evidence 现在同时生成并校验 `release/manifest/latest.json.sha256`，确保发布 manifest 和 checksum artifact 成对存在。
- `release/manifest/template.json` 将 checksum 纳入必需 artifacts，发布产物清单能完整描述可验证 Evidence。

### 治理

- `make release-check`、`make release-check-extended` 和 `make release-final-check` 强制要求 `GOWORK=off`，避免发布门禁受外部 workspace 污染。
- `docs-check` 新增标准源、checksum artifact、`GOWORK=off` 发布命令和 fuzz smoke 文档约束，防止文档与发布 Harness 漂移。
- 扩展 `golangci-lint` 规则集，提高模板基础库的静态质量门槛。

## v0.2.0 - 2026-06-01

### 新增

- 新增 `make release-preflight VERSION=vX.Y.Z`，在打 tag 前检查版本、`main` 同步状态、目标 tag、`CHANGELOG.md`、必需工具和最终 release gate。

### 修复

- Release Check workflow 在运行 `make release-check` 前安装 `golangci-lint` 和 `govulncheck`，并使用 `GOWORK=off`，与 CI 的强制 gate 环境保持一致。
- Release Evidence 校验新增目标版本比对，避免目标 tag 与 `manifest.version` 不一致。

## v0.1.0 - 2026-06-01

### 新增

- 初始化 `baselib-template` 结构。
- 添加标准 Go 基础库包骨架。
- 添加 Makefile 命令。
- 添加 Harness Gate 脚本。
- 添加 GitHub Actions 工作流。
- 添加 contracts 文件。
- 添加 Agent 运行时模板。
- 添加 release manifest 模板。
- 添加 typed error、错误包装和 `ErrorKind` contract。
- 添加 client 生命周期、健康检查和请求扩展 metrics contract。
- 添加 health JSON contract 与 contracts 回归测试。
- 添加 config schema 到 `Config` 字段映射的 contract 回归测试。
- 添加 `scripts/render_template.sh`，支持生成 `foundationx` 等具体基础库。
- 添加 `examples/basic`、`examples/config` 和 `examples/health` smoke 测试，锁定文档示例输出。
- 添加 `testkit` 夹具和断言回归测试。
- 添加配置属性测试、配置 fuzz smoke 测试、健康状态 golden 测试和 `testkit` golden 文件工具。

### 安全

- 添加 Secret Gate。
- `make security` 强制运行 `govulncheck ./...` 和密钥扫描；缺少 `govulncheck` 时必须失败。
- 配置脱敏规则覆盖 release Evidence 和日志可见内容。
- Boundary Gate 同时拦截 `github.com/bytechainx/x.go` 和 `github.com/ZoneCNH/x.go`。

### 治理

- 添加 Evidence 和复盘模板。
- CI 在 `make ci` 前安装 `golangci-lint` 和 `govulncheck`，与 Makefile 强制 gate 对齐。
- `make release-check` 统一执行 CI、integration 和 manifest 生成。
- `make release-final-check` 在发布前串联 `release-check`、release Evidence 校验和工作区洁净校验。
- `make integration` 通过临时 `foundationx` 和 `corekit` 渲染、测试、contracts、boundary 与 Evidence 生成验证模板链路。
- `release/manifest/latest.json` 作为生成产物保留在源码历史之外，避免 release Evidence 与源码提交互相污染。

### 验证

- 发布前已运行 `GOWORK=off make release-final-check`。
- `go fmt ./...`、`go vet ./...`、`golangci-lint run ./...`、`go test ./...`、`go test -race ./...`、Boundary、Security、contracts、integration 和 release Evidence 校验均通过。
- `v0.1.0` 为 annotated tag，指向提交 `b6dfe9b93e4417a3b7e077cec1b4c0fffdc37240`。
