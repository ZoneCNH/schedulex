# 发布模板

## 占位符

- `{{MODULE_NAME}}`
- `{{MODULE_PATH}}`
- `{{PACKAGE_NAME}}`

## Release Gate

- `make ci`
- `make integration`
- `make evidence`
- `make release-evidence-check`

推荐入口是：

```bash
XLIB_CONTEXT=release_verify GOWORK=off make release-check
```

`GOWORK=off` 用于证明模板不依赖父级 workspace。Makefile 的 gate 入口统一通过 `cmd/goalcli` 调度；shell 脚本仍保留为兼容实现层，供本地排障和旧自动化复用。

Full Goal Runtime v3.1 的评分入口是：

```bash
GOWORK=off go run ./cmd/goalcli score --min 9.8
```

CI 和 release workflow 必须在 release gate 后执行该评分，防止 Makefile、CI、文档和下游 integration 的契约漂移。

发布前的最终入口是：

```bash
XLIB_CONTEXT=release_verify GOWORK=off make release-final-check
```

`release-final-check` 会在完整 gate 之后要求 `release/manifest/latest.json` 与当前 HEAD、源码摘要、contract 指纹和依赖清单一致，并要求 git 工作区为 `clean`。它适合在打 tag 或发布前运行；开发中的 `release-check` 允许工作区因为未提交改动显示为 `dirty`，但仍会校验 manifest 与当前内容一致。

打 tag 前推荐使用 release preflight：

```bash
XLIB_CONTEXT=release_verify GOWORK=off make release-preflight VERSION=v0.4.6
```

`release-preflight` 会先检查版本号、当前分支、工作区洁净状态、`main` 与 `origin/main` 是否一致、目标 tag 是否已存在、`CHANGELOG.md` 是否包含目标版本，以及 `golangci-lint` / `govulncheck` 是否已安装；随后以 `GOWORK=off` 和 `XLIB_CONTEXT=release_verify` 运行 `release-final-check`。tag 应在该入口通过后再创建和推送。

## GitHub Release 发布对象

推送 `v*` tag 后，`.github/workflows/release.yml` 必须在 `release-final-check` 通过后自动创建或更新同名 GitHub Release。workflow 使用 `gh release create` / `gh release edit` 发布，并用 `gh release view` 校验 Release 对象存在、不是 draft、不是 prerelease。只有 tag 而没有 GitHub Release 对象时，发布视为未完成。

## Required Release Check

`make release-check` 是默认发布门禁，必须通过：

```text
ci
integration
evidence
release-evidence-check
```

## Extended Release Check

`make release-check-extended` 是发布前强验证，推荐在重要版本、公共 API 变更、contract 变更、schema 变更、metrics 变更时执行：

```text
ci-extended
integration
evidence
release-evidence-check
```

`make ci-extended` 会在默认 `ci` 外追加：

```text
property
golden
fuzz-smoke
```

## Gate 工具契约

`make ci` 中的 `make lint` 和 `make security` 是强制 gate。运行前必须可用：

- `golangci-lint`
- `govulncheck`

缺少任一工具时，本地 Makefile 必须硬失败。GitHub Actions CI 和 Release Check workflow 会在运行 `make ci` / `make release-check` 前安装 `golangci-lint` 和 `govulncheck`，以保证本地与远端 workflow 对同一组强制 gate 负责。

GitHub Actions workflow 引用的第三方 Action 必须固定为 40 位 commit SHA，并用注释保留来源 tag 供审计。CI、Release Check 和 Security workflow 安装 `govulncheck` 时必须使用固定版本；当前基线是 `golang.org/x/vuln/cmd/govulncheck@v1.3.0`，不得在发布门禁中使用 `@latest`。

`make security` 必须同时运行 `govulncheck ./...` 和 `scripts/check_secrets.sh`；不得把漏洞扫描降级为可选检查。

## Evidence

发布 Evidence 生成到 `release/manifest/latest.json`，该文件是生成产物，不提交到源码历史。提交到仓库的是 `release/manifest/template.json`；CI release workflow 会上传 `latest.json` 作为 artifact。

`latest.json` 至少包含：

- `module`
- `version`
- `commit`
- `tree_sha`
- `source_digest`
- `tracked_file_count`
- `go_version`
- `generated_at`
- `generated_by`
- `tree_state`
- `checks`
- `contracts`
- `dependencies`
- `tools`
- `standard_impact`
- `downstream_sync_required`
- `generator_evidence`
- `workflow`
- `score`
- `artifacts`
- `notes`

其中 `standard_impact.downstream_release_decision` 只能使用 `required` 或 `not_required`；`standard_impact.repository_rules_release_decision` 只能使用 `audit_required` 或 `not_required`。release manifest 校验必须拒绝其它非空值，空值仍按 required field 处理。

`make release-check` 成功后会以 `CHECK_STATUS=passed` 生成 manifest，并立即运行 `make release-evidence-check`。若单独运行 `make evidence`，未显式传入的检查状态默认为 `unknown`，后续校验会拒绝把这些状态当作已通过的 release gate。因为 `latest.json` 不再提交，manifest 中的 `commit` 可以指向实际执行 release gate 的 HEAD，避免自引用提交哈希导致的永久漂移。

Extended Evidence 推荐额外记录：

- `make ci-extended` 结果。
- `make property` 结果。
- `make fuzz-smoke` 结果。
- `make golden` 结果。
- compatibility 和 observability contract 结果。

`source_digest` 基于 `git ls-files` 中的受跟踪文件内容计算；`contracts` 固定记录核心 contract 文件的 SHA256；`dependencies` 来自 `go list -m -json all`；`tools` 记录 Go、`golangci-lint` 和 `govulncheck` 的版本或可用状态。这些字段由 `internal/tools/releasemanifest` 生成并校验，不再由 shell 拼接 JSON。

`make integration` 会通过 `cmd/goalcli integration` 调用 `scripts/render_template.sh`，生成临时 `kernel` 和 `corekit` 两个下游库，并对每个生成目录执行：

- 模块路径、包目录和旧模板标识扫描。
- `GOWORK=off go test ./...`
- `GOWORK=off make contracts`
- `GOWORK=off make boundary`
- `CHECK_STATUS=passed GOWORK=off make evidence`
- `RELEASE_EVIDENCE_REQUIRE_PASSED=1 GOWORK=off make release-evidence-check`

这一步用于证明模板替换、包目录迁移、imports、contracts、边界检查和生成后 release Evidence 都能在下游库中独立工作。

## 规则

- 没有 Evidence artifact 不得发布。
- `tree_state` 为 `dirty` 时可以在开发中生成 Evidence，但正式发布前必须通过 `make release-final-check`。
- 不得在 release manifest、PR、Issue 或变更日志条目中包含原始凭据。
- 不得依赖 `github.com/bytechainx/x.go` 或 `github.com/ZoneCNH/x.go`。
- public API、config schema、error kind、health JSON、metrics name 或 release manifest schema 变更必须在 release notes 或 release manifest 中显式标记 breaking change。

## Release Score 与 Workflow Evidence

发布分数是 release gate 的显式合同：

```bash
go run ./cmd/goalcli score --min 9.8
```

`release/manifest/latest.json` 必须记录 `score` 与 `workflow` 字段。`workflow_run_id`、`artifact_name`、`artifact_url` 用来把本地 manifest 与 GitHub Actions 上传的 `release-manifest-<workflow-run-id>` artifact 对齐；本地运行时允许使用 `local:*` artifact URL。`release-final-check` 会在 clean tree 要求之外校验 manifest 内的 score threshold。

## Debt evidence

Release checks generate debt evidence with `make debt-evidence` before manifest generation. `release-final-check` enforces `goalcli debt --mode enforce --min-score 9.8` and release evidence verification validates the manifest `debt` block. Generated `release/debt/*` artifacts are not committed.
