# 供应链与 Evidence

## 目标

本模板的 release Evidence 不是普通构建日志，而是可重复校验的发布事实清单。它必须回答三个问题：

- 当前发布对应哪个 Go module、commit 和 git tree。
- 当前源码、contract 文件和依赖清单是否与 manifest 一致。
- 必需 gate 是否全部以 `passed` 状态完成。

## Manifest 生成

`make evidence` 调用 `scripts/generate_manifest.sh`，最终由 `internal/tools/releasemanifest` 生成 `release/manifest/latest.json`。生成内容包括：

- `commit` 和 `tree_sha`：来自当前 git HEAD。
- `source_digest` 和 `tracked_file_count`：来自 `git ls-files` 中所有受跟踪文件的路径和内容摘要。
- `contracts`：核心 contract 文件的 SHA256 指纹。
- `dependencies`：`go list -m -json all` 的模块清单。
- `tools`：Go、`golangci-lint` 和 `govulncheck` 的版本或可用状态。
- `checks`：`fmt`、`vet`、`lint`、测试、race、boundary、secret scan、security、contract 和 integration gate 状态。

`release/manifest/latest.json` 是生成产物，不提交源码历史；`release/manifest/template.json` 只保留字段模板。

## Manifest 校验

`make release-evidence-check` 会重新读取当前仓库事实，并校验：

- manifest 的 module、commit、tree SHA、源码摘要和受跟踪文件数量与当前仓库一致。
- contract 指纹和依赖清单与当前文件、当前 Go module 解析结果一致。
- 必需 check 均存在，且在 release gate 中必须为 `passed`。
- artifact 列表包含 `release/manifest/latest.json` 和 `release/manifest/latest.json.sha256`。

`make release-final-check` 在上述校验之外要求 `tree_state=clean`。正式发布、打 tag 或交付给下游基础库前必须使用该入口。

## CI Artifact

GitHub Actions 运行 `GOWORK=off make release-check`，并上传 `release/manifest/latest.json` 与 `release/manifest/latest.json.sha256` 作为 `release-manifest-<workflow-run-id>` artifact。CI 中上传的 artifact 是发布 Evidence 的外部留痕；本地生成的 `latest.json` 和 checksum 用于验证和排障。

## Workflow 供应链固定

GitHub Actions workflow 必须使用 40 位 commit SHA 固定第三方 Action，并在同一行用注释记录来源 tag，例如 `# tag v4.2.2`。不得使用 `actions/checkout@v4`、`actions/setup-go@v5` 这类浮动 tag 作为最终发布门禁配置；版本更新应通过维护者审查或 dependabot PR 重新 pin 到新的 commit。

CI、Release Check 和 Security workflow 安装 `govulncheck` 时必须使用固定版本。当前基线是 `golang.org/x/vuln/cmd/govulncheck@v1.3.0`；升级时应同步更新 workflow、发布文档和验证记录。

Release manifest 相关测试必须在临时 fixture 仓库内构造所需 `.omc` state 文件，不得依赖当前工作区的 Agent 运行态。这样可以保证 Evidence 测试在 clean checkout、CI 和本地开发目录中行为一致。

## 下游模板安全线

`make integration` 会渲染 `kernel` 和 `corekit` 两个临时下游库，检查旧模板标识是否清空，并在下游库内生成、校验 release Evidence。这保证模板替换逻辑、contract gate、boundary gate 和 Evidence 工具不会只在模板仓库自身成立。旧 `foundationx` 只作为迁移兼容扫描项，不再作为默认下游。

## Score 与 Workflow Evidence

供应链 Evidence 必须包含可机器校验的 `score` 与 workflow artifact 元数据。`score` 来自 `go run ./cmd/goalcli score --min 9.8`，用于汇总 manifest schema、release gate、security scan、scorecard 文档、release/retrospective 模板等发布质量维度。`workflow_run_id`、`artifact_name`、`artifact_url` 用于追踪 CI 上传的 `release/manifest/latest.json` 与 `release/manifest/latest.json.sha256`，防止只有本地文件而没有外部 artifact 留痕。

## Dependency debt purpose

New runtime dependencies require an approved purpose in `.agent/debt/dependency-purpose.yaml` or an ADR. The dependency debt gate flags unreviewed installer patterns and feeds the release debt evidence manifest block.
