# Downstream Compatibility

`xlib-standard` 的下游兼容性必须通过真实生成库 smoke 证明，而不是只证明模板仓库自身可用。

## 默认下游

- `kernel`：默认 L0 集成目标，module path 为 `github.com/ZoneCNH/kernel`，是 Full Goal Runtime v3.1 的必跑下游集成目标。
- `corekit`：中性组织路径 smoke，module path 为 `example.com/acme/corekit`，用于证明 generator 不依赖固定组织、GitHub owner 或 module prefix。

旧 `foundationx` 只作为迁移兼容名出现，不再是默认下游。

## 目标库矩阵

详细 module path、package、layer、allowed deps 和 forbidden deps 见 [`../downstream-matrix.md`](../downstream-matrix.md)。矩阵至少覆盖：`kernel`、`configx`、`observex`、`testkitx`、`postgresx`、`redisx`、`kafkax`、`taosx`、`ossx`、`clickhousex`。

## 工具要求

| 工具 | 用途 | 要求 |
| --- | --- | --- |
| Go 1.23 | 编译、测试、`go mod tidy`、dependency list | 本地和 CI 一致 |
| `make` | 执行 Harness gate | 必须可运行 required targets |
| `git` | 初始化临时下游、检查 clean diff、计算 commit/tree | integration 和 Evidence 必需 |
| `golangci-lint` | `make lint` | 缺失时必须失败 |
| `govulncheck` | `make security` | 缺失时必须失败 |
| `python3` | docs link checker | `make docs-check` 必需 |
| `sha256sum` | 计算 `latest.json` hash | CI artifact Evidence 必需 |
| GitHub Actions artifact | 保存 `release/manifest/latest.json` | 远端 release Evidence 必需 |

## Gate

`GOWORK=off make integration` 是默认下游兼容 gate。它通过 `cmd/goalcli integration` 覆盖 generator smoke、`kernel`/`corekit` 代表路径和关键边界检查。

生成出的每个代表下游必须通过：

```bash
GOWORK=off go mod tidy
GOWORK=off go test ./...
GOWORK=off make contracts
GOWORK=off make boundary
CHECK_STATUS=passed GOWORK=off make evidence
RELEASE_EVIDENCE_REQUIRE_PASSED=1 GOWORK=off make release-evidence-check
```

这些命令需要在 `kernel` 和 `corekit` 的渲染结果中通过；失败时不得宣称 downstream compatible。当新增 profile 时，在不污染默认 `make ci` 的前提下补充 profile-specific smoke 或 extended gate。

## 兼容破坏

以下情况视为兼容破坏：

- 删除 generator 必需输入。
- 破坏 `Config`、`New`、`Close`、`HealthCheck` 的基础语义。
- 移除 required gate。
- 让生成库依赖 `x.go`。
- 让生成库不能独立生成 Evidence。
