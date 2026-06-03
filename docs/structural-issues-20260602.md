# xlib-standard 结构性问题分析

> 生成日期：2026-06-02
> 基于：深度代码审查、治理文件审计、Makefile 依赖图分析

---

## 当前阅读说明

本文保留 2026-06-02 审计时的结构性问题快照。部分条目已经在后续治理中修复；阅读时以文末“当前修复状态”表为准。特别是版本号体系已在 v0.4.3 口径下统一，本文正文中的 `v0.3.7` / `v0.1.0` 描述仅代表历史问题状态。

## 问题 1：`cmd/goalcli` 超级文件与上帝 switch

**位置**：`cmd/goalcli/governance.go`（743 行、40 个函数）、`cmd/goalcli/main.go`（213 行、29 个 case 的 switch）

**描述**：所有门禁逻辑集中在一个文件中。`main.go` 的 `run()` 函数将命令分为三类：Go 原生实现（score、version、doctor）、委托 shell 脚本（boundary、contracts、docs-check）、planned command（仅检查文件是否存在）。

**影响**：

- 新增门禁需要同时改 `main.go` 的 switch、`governance.go` 的 `plannedCommandFiles` map、`command-registry.yaml`、`issue-registry.yaml`、`Makefile`、`makefile-baseline.yaml`、`makefile-target-registry.yaml` —— **7 个文件联动**
- `runPlannedCommand` 本质上只是检查 YAML 中声明的文件是否存在于磁盘，不执行任何真实逻辑，但通过了所有 gate 验证

**根因**：没有按职责拆分（guard 命令、registry 检查、planned 命令、外部委托应该分文件/分包）

**建议**：按职责拆分为 `guards.go`、`registry.go`、`planned.go`、`external.go`，或将 planned command 的验证逻辑下沉到各 YAML 对应的实现中。

---

## 问题 2：Planned Command 的空壳验证

**位置**：`cmd/goalcli/governance.go:398-434`（`plannedCommandFiles` map）、`runPlannedCommand` 函数

**描述**：`plannedCommandFiles` map 中约 30 个命令（占总命令一半）只检查声明的 YAML 文件是否存在，不执行任何实质性验证。例如 `agent-team-contract` 只检查 `.agent/team-contract.yaml` 是否存在。

**影响**：

- 这些命令在 `--verify` 模式下返回 `passed`，但实际上只证明了"文件存在"，不证明"契约被满足"
- `p1-governance-check` 和 `p2-runtime-check` 的 36 个子命令全部走这条路径，gate 密度高但验证深度浅
- issue-registry 中标记为 `status: implemented` 的 issue 可能只是"文件占位"而非"功能实现"

**建议**：为每个 planned command 定义最小可验证断言（至少检查 YAML 内容的关键字段），而非仅检查文件存在性。

---

## 问题 3：版本号体系混乱（历史快照，当前已修复）

**位置**：多处

以下版本关系为 2026-06-02 审计时的历史状态。当前 release 口径已将项目发布版本、`templatex.Version`、release manifest template、release preflight 文档和 harness 配置同步到 `v0.4.3`。

| 版本     | 位置                                         | 含义                     |
| -------- | -------------------------------------------- | ------------------------ |
| `v0.3.7` | CHANGELOG.md                                 | 项目发布版本             |
| `v0.1.0` | `pkg/templatex/version.go`                   | 包版本（**未同步**）     |
| `v2.9.3` | `cmd/goalcli/governance.go` goalcliVersion | Goal Runtime 版本        |
| `v3.1`   | `.agent/goal-runtime.md`                     | Goal Runtime schema 版本 |

**影响**：

- 历史上 `templatex.Version` 停留在 `v0.1.0`，而项目已发布到 `v0.3.7`
- 历史上 `release-preflight` 的 `VERSION=v0.2.0` 是硬编码的过期值
- 当前 v0.4.3 已修复该漂移；后续风险转为“新增发布时必须同步更新版本事实”

**建议**：统一版本管理策略，`templatex.Version` 应与 CHANGELOG 保持同步，或通过 build info 注入。

---

## 问题 4：Issue Registry 硬编码计数

**位置**：`cmd/goalcli/governance.go:675-688`

```go
func requiredIssueRegistryNeedles() []string {
    // ...
    {prefix: "P0", count: 18},
    {prefix: "P1", count: 21},
    {prefix: "P2", count: 15},
    {prefix: "CTX", count: 4},
}
```

**描述**：Go 代码中硬编码了 P0=18、P1=21、P2=15、CTX=4 的 issue 数量。gate 验证时按这些数字生成 needle（如 `P0-018`），然后检查 YAML 是否包含。

**影响**：

- 新增 issue 必须同步改 Go 代码中的 `count` 值，否则 gate 会要求 YAML 包含不存在的 issue ID
- 反过来，如果删除了 issue 但没改 count，gate 仍然通过（因为 needle 存在但可能已无意义）
- 这是典型的**配置硬编码反模式**

**建议**：从 YAML 动态解析 issue 数量，或在 YAML 中声明 expected count。

---

## 问题 5：Shell 脚本与 Go 代码的职责分裂

**位置**：`scripts/` 目录（14 个脚本）、`cmd/goalcli/main.go`（`runExternal` 函数）

**描述**：14 个 shell 脚本和 Go 代码各自承担一部分 gate 逻辑，但边界不清：

- `check_docs.sh` 是 **470+ 行**的 Python+Shell 混合脚本，包含文档存在性检查、文本 needle 检查、链接完整性检查、模板占位符检查、Makefile 结构检查、命名漂移检查
- `check_boundary.sh` 做 x.go 依赖检查和业务术语过滤
- `check_secrets.sh` 做密钥扫描

但 `goalcli` CLI 通过 `runExternal` 只是委托执行，不做任何包装或结果解析。

**影响**：

- Shell 脚本的输出格式（纯文本）与 Go gate 的输出格式（JSON `gateReport`）不统一
- Shell 脚本的错误无法被 `goalcli` 结构化捕获和聚合
- `check_docs.sh` 混用 Python3 heredoc 和 Shell，调试困难
- `check_docs.sh` 中的 70+ 条 `require_text` 断言是字符串包含检查，文档措辞变更即可触发 gate 失败

**建议**：将 shell 逻辑逐步迁入 Go，或至少让 shell 脚本输出 JSON 格式的 gateReport。

---

## 问题 6：`.agent/` 目录膨胀与治理文件冗余

**位置**：`.agent/` 目录（80+ 文件）

**描述**：大量文件是小体积 YAML（< 200 字节），且许多只是声明一个文件路径或一个布尔值：

```
.agent/runtime-health.yaml      (56 bytes)
.agent/runtime-install.yaml     (56 bytes)
.agent/upgrade-runtime.md       (100 bytes)
.agent/self-healing-skeleton.md (116 bytes)
```

**影响**：

- 这些文件的存在意义主要是为了让 `plannedCommandFiles` 检查通过，而非承载独立信息
- `.agent/` 同时包含**运行时工件**（harness.yaml、goal-runtime.md）和**占位文件**（仅为满足 gate），读者无法区分哪些是"活的"哪些是"死的"
- 治理文件数量增长到难以人工审计的程度

**建议**：清理占位文件，引入元数据字段区分 `status: placeholder` vs `status: active`。

---

## 问题 7：docs-check 的脆弱性

**位置**：`scripts/check_docs.sh`（470+ 行）

**描述**：`check_docs.sh` 包含：

- 70+ 条 `require_text` 断言（检查特定文件是否包含特定字符串）
- 内嵌 Python3 脚本做正则匹配和链接检查
- Makefile 结构解析（用正则而非 AST）

**影响**：

- 任何文档措辞变更都可能触发 gate 失败（例如把"kernel"改成"Kernel"）
- 断言是**字符串包含**而非**语义匹配**，导致文档为了通过 gate 而堆砌关键词
- Python 脚本内嵌在 Shell heredoc 中，IDE 无法提供语法高亮和类型检查

**建议**：将文本断言迁入 Go 代码（利用已有的 `runRegistryCheck` 模式），或引入结构化断言格式（YAML 声明式 needle 检查）。

---

## 问题 8：测试代码量倒挂

**位置**：`internal/tools/releasemanifest/main_test.go`（1375 行）、`cmd/goalcli/main_test.go`（1090 行）

**描述**：

| 文件                           | 行数 | 被测文件行数 | 比值  |
| ------------------------------ | ---- | ------------ | ----- |
| `releasemanifest/main_test.go` | 1375 | 775          | 1.78x |
| `cmd/goalcli/main_test.go`    | 1090 | 956          | 1.14x |

**影响**：

- 测试代码维护成本高于实现代码
- 测试本身成为另一个需要维护的"系统"
- `releasemanifest` 测试跑 **38 秒**，拖慢 CI 反馈循环

**建议**：拆分大测试文件，引入 table-driven 测试减少重复，对慢测试做 profiling 优化。

---

## 问题 9：Context Profile DAG 与 Makefile 的循环依赖

**位置**：Makefile 中 `release-final-check`、`context-release`、`context-full` 的依赖关系

**描述**：

```
release-final-check → context-release → context-full → governance-check + p1-governance-check + p2-runtime-check
                    → score --min 9.8
                    → check_release_evidence.sh
```

`context-release` 的实现是 `CHECK_STATUS=passed $(MAKE) evidence` + hash + check，而 `release-final-check` 又调用 `$(MAKE) context-release`。

**影响**：

- `release-final-check` 实际上会跑两遍部分 gate（一遍在 `context-release` 中，一遍在自身的 `RELEASE_EVIDENCE_REQUIRE_PASSED=1` 检查中）
- `context-profile-check` 验证 Makefile 结构，但 Makefile 本身又通过 `$(GOALCLI)` 调用 `context-profile-check` —— 形成间接自引用

**建议**：文档化完整依赖图，消除重复 gate 执行，或将 `context-release` 的 evidence 生成步骤提取为独立 target。

---

## 问题 10：下游生成缺乏端到端验证

**位置**：`scripts/render_template.sh`、`scripts/run_integration.sh`

**描述**：`render_template.sh` 声称可以生成 10 个下游库（kernel、configx、observex 等），但：

- `make integration` 只跑一个 `run_integration.sh`，未验证所有下游
- 没有在 CI 中实际渲染并测试下游库
- `downstream-matrix.md` 记录了兼容矩阵，但数据来源不明

**影响**：

- Generator 的正确性无法在本仓库内闭环验证
- 下游库的 breaking change 只能靠人工发现

**建议**：在 CI 中添加至少一个下游库（如 kernel）的渲染 + 编译 + 测试流水线。

---

## 严重度排序

| 严重度 | 问题                                              | 修复难度                           |
| ------ | ------------------------------------------------- | ---------------------------------- |
| 🔴 高  | Planned Command 空壳验证（gate 通过但无实质验证） | 需要重新设计 P1/P2 gate 的验证策略 |
| 🔴 高  | Issue Registry 硬编码计数                         | 改为动态解析 YAML                  |
| 🟡 中  | 版本号体系混乱（4 套版本不同步）                  | 统一版本管理策略                   |
| 🟡 中  | governance.go 超级文件                            | 按职责拆分文件                     |
| 🟡 中  | Shell/Go 职责分裂 + 输出格式不统一                | 将 shell 逻辑逐步迁入 Go           |
| 🟡 中  | docs-check 脆弱性（字符串包含断言）               | 引入结构化断言或 AST 解析          |
| 🟢 低  | .agent/ 目录膨胀                                  | 清理占位文件，区分运行时 vs 占位   |
| 🟢 低  | 测试代码量倒挂 + releasemanifest 慢测试           | 拆分测试、引入 test caching        |
| 🟢 低  | Context Profile 与 Makefile 循环依赖              | 文档化依赖图，消除重复 gate        |
| 🟢 低  | 下游生成缺乏端到端验证                            | CI 中添加下游渲染+测试             |

---

## 2026-06-03 修复执行记录

本轮使用 agent team fanout 对问题清单做并行复核后，优先修复可闭环验证的结构问题 2、3、4，保留问题 1、5、7 等大范围拆分/迁移项作为后续专项。

| 问题 | 状态 | 修复摘要 | 验证 |
| ---- | ---- | -------- | ---- |
| 问题 2：Planned Command 空壳验证 | 已修复最小语义层 | `runPlannedCommand` 不再只检查文件存在；会拒绝目录、空文件、非法 JSON，并对 `agent-team-contract`、`runtime-health`、`execution-context` 等核心 manifest 检查最小语义 marker。 | `GOWORK=off go run ./cmd/goalcli agent-team-contract --dry-run --verify`、`runtime-health --dry-run --verify`、`execution-context --dry-run --verify` |
| 问题 3：版本号体系混乱 | 已修复当前 release 口径 | 拆分项目发布版本 `projectReleaseVersion` 与治理运行时版本 `governanceRuntimeVersion`，并将 `templatex.Version`、release manifest template、release preflight 文档和 harness 版本同步到 `CHANGELOG.md` 最新版本 `v0.4.3`。 | `GOWORK=off go run ./cmd/goalcli version --json`、`GOWORK=off go test ./cmd/goalcli` |
| 问题 4：Issue Registry 硬编码计数 | 已修复 | 移除 Go 代码中的固定 P0/P1/P2/CTX 数量 needle，改为从 `.agent/issue-registry.yaml` 动态解析 issue ID，校验 ID 格式、重复、连续性、`status: implemented`、`command` 和非空 `evidence`。 | `GOWORK=off go run ./cmd/goalcli issue-registry`、`GOWORK=off go run ./cmd/goalcli context-profile-check` |

回归验证：`GOWORK=off go test ./...` 已通过。
