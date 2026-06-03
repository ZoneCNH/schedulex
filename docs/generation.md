# 生成模板

## 用途

`scripts/render_template.sh` 用于把 `xlib-standard` 参考模板渲染为具体基础库，例如 `kernel`。标准源仓库是 [`xlib-standard`](https://github.com/ZoneCNH/xlib-standard)，并同时承载模板、generator、Harness 和 Evidence 实现。旧 `baselib-template` / `foundationx` 名称只作为迁移文档语境保留。脚本负责同步替换 module name、module path、package name、`pkg/` 目录名、imports、文档占位符和脚本中的模板名称。

## 示例

```bash
scripts/render_template.sh \
  --module-name kernel \
  --module-path github.com/ZoneCNH/kernel \
  --package-name kernel \
  --out ../kernel
```

`--out` 必须指向不存在或为空的目录，避免覆盖已有仓库内容。

## 渲染范围

- `{{MODULE_NAME}}` 替换为 `--module-name`。
- `{{MODULE_PATH}}`、`github.com/ZoneCNH/xlib-standard` 和迁移兼容的 `github.com/ZoneCNH/baselib-template` 替换为 `--module-path`。
- `xlib-standard` 和迁移兼容的 `baselib-template` 替换为 `--module-name`。
- `{{PACKAGE_NAME}}`、`pkg/templatex` 和 `templatex` imports 替换为 `--package-name`。
- 文档、Go 代码、JSON contract、shell 脚本、Makefile 和 CI 配置同步更新；标准源仓库仍是 [`https://github.com/ZoneCNH/xlib-standard`](https://github.com/ZoneCNH/xlib-standard)，渲染产物中的源身份会改写为下游 module identity，避免残留模板仓库名称。

脚本不会复制 `.git`、`.omx`、`.worktree` 和 `release/manifest/latest.json`。`latest.json` 是生成产物，生成后的库必须自己运行 release gate 生成新的 Evidence artifact。

## 验证

生成后至少运行：

```bash
GOWORK=off make release-check
```

模板自身的 `make integration` 会渲染两个临时下游库：

- `kernel`：目标仓库路径 `github.com/ZoneCNH/kernel`，用于证明默认 L0 下游目标仍可生成。
- `corekit`：中性路径 `example.com/acme/corekit`，用于证明替换逻辑不依赖特定组织或包名。

每个临时库都会运行以下验证：

- `scripts/check_rendered_template.sh`：确认 `go.mod` module path、`pkg/<package>` 目录、旧模板目录、旧 module path、占位符和 `templatex` 标识。
- `GOWORK=off go mod tidy` 后检查 `go.mod` / `go.sum` 没有未提交差异。
- `GOWORK=off go test ./...`
- `GOWORK=off make contracts`
- `GOWORK=off make boundary`
- `CHECK_STATUS=passed GOWORK=off make evidence`
- `RELEASE_EVIDENCE_REQUIRE_PASSED=1 GOWORK=off make release-evidence-check`

这组验证用于防止生成脚本、包路径、imports、contract gate、boundary gate 和生成后 Evidence 回归。

## 生成后 Release Evidence

生成后的库会继承 `internal/tools/releasemanifest`。该工具会生成并校验 `release/manifest/latest.json`，其中包括当前 HEAD、tree SHA、源码摘要、contract SHA256、依赖清单和工具版本。发布前应使用：

```bash
GOWORK=off make release-final-check
```

`release-final-check` 要求所有 gate 状态为 `passed`，并要求 git 工作区为 `clean`。如果只是开发中自测，`make release-check` 已足够；它允许工作区显示 `dirty`，但仍会验证 manifest 和当前源码内容一致。

## 边界

生成后的基础库仍必须保持独立，不能依赖 `github.com/bytechainx/x.go`、`github.com/ZoneCNH/x.go` 或任何 `x.go/internal/*` 包；标准规则继续引用独立仓库 [`xlib-standard`](https://github.com/ZoneCNH/xlib-standard)。
