# 发布标准

`schedulex` v1.0.0 发布必须证明源码、公共 API、调度器契约、release manifest、CI 状态和 GitHub Release 对象一致。tag 存在但没有 GitHub Release 对象时，发布视为未完成。

## 发布路径

1. 在 clean workspace 运行 `GOWORK=off make release-check VERSION=v1.0.0`。
2. 确认 `release/manifest/latest.json` 和 `release/manifest/latest.json.sha256` 由 `GOWORK=off make evidence` 生成，且 checksum 校验通过。
3. 确认 `GOWORK=off make score` 达到 `score >= 9.8`。
4. 推送版本 tag，由 `.github/workflows/release.yml` 运行 release gate 并上传 manifest artifact。
5. release workflow 使用 `gh release create` 或 `gh release edit` 发布同名 GitHub Release，并立即用 `gh release view` 校验对象存在、非 draft、非 prerelease。
6. 在 PR 或 release notes 中附上已运行命令、manifest artifact、checksum、CI run 和 known gaps。

## 命令

```bash
GOWORK=off make release-check VERSION=v1.0.0
GOWORK=off make release-preflight VERSION=v1.0.0
GOWORK=off make release-final-check
GOWORK=off make score
```

`release-check` 是发布入口，覆盖 `ci-extended` 和 `release-preflight`。`ci-extended` 覆盖格式化、vet、lint、测试、race、边界、契约、文档、安全、调度器专项检查、evidence、governance、P1、P2、final check 和 score。`release-preflight` 必须显式传入 `VERSION`，并与 tag、release notes 和 manifest 保持一致。

## Manifest

`release/manifest/latest.json` 是生成产物：

- 可以作为 CI artifact 上传。
- 可以作为本地 Evidence 检查输入。
- 不提交到源码历史。
- `release/manifest/latest.json.sha256` 是对应 checksum 产物，随 CI artifact 上传，并保持在 `.gitignore` 中。
- manifest 必须记录 `score` 和 `workflow`；`workflow_run_id`、`artifact_name`、`artifact_url` 用于对齐 CI 上传的 release manifest artifact，本地运行时可使用 `local:*` Evidence URL。

Release manifest 相关测试必须在临时 fixture 仓库构造所需状态，不得依赖当前工作区的 Agent 运行态文件。

## CI 与保护规则

主分支保护必须要求以下 status checks：

- `ci`
- `release-check`
- `security`
- `integration`
- `gates`
- `worktree-check`

`.github/workflows/ci.yml` 必须包含 `make release-check` 并固定安装 `golangci-lint` 和 `govulncheck`。`.github/workflows/goal-gates.yml` 必须在 pull request、main push 和手动触发时运行 `GOWORK=off make ci-extended VERSION=v1.0.0`。

## 供应链约束

- GitHub Actions workflow 引用的第三方 Action 必须固定为 40 位 commit SHA，并在同一行保留来源 tag 注释。
- CI、Release Check、Goal Gates、Integration、Security 和 Release workflow 安装 `govulncheck` 时必须使用固定版本；当前基线是 `golang.org/x/vuln/cmd/govulncheck@v1.3.0`。
- 本地缺少 `golangci-lint` 或 `govulncheck` 时，`make lint` / `make security` 必须失败，不得把必需 gate 记录为跳过。

## 版本与发布对象

- `VERSION` 必须显式传入 `release-check` 或 `release-preflight`。
- tag、release notes、manifest、GitHub Release 名称和 artifact 名称必须一致。
- 工作区 dirty、tag 未推送或 GitHub Release 对象缺失时，不得宣称最终发布完成。
- Release workflow 必须声明 `contents: write`，并对发布命令使用 `--verify-tag`，保证 Release 对象只能绑定到已存在的远端 tag。
