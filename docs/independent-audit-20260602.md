# xlib-standard 独立审计报告

> 审计日期：2026-06-02
> 审计方式：Agent Team 四维度并行独立评审（架构 / 测试 / 文档治理 / 安全供应链）
> 审计范围：38 个 .go 文件（约 3632 行）、4 个 CI workflow、35 个 .agent 工件、docs/ 全量标准文档
> 基线：项目自带 `goalcli score` 自评 **10.0 / 10**（阈值 9.8，状态 passed）
> 修复对齐：2026-06-02 已关闭 C1、C2、C3、H2、H3；原始评分仍作为审计时快照保留，状态见第六节。

---

## 一、执行摘要

**独立综合评分：7.1 / 10**

xlib-standard 的公共层（`pkg/templatex`）工程质量上乘——API 设计规范、错误模型与并发控制符合 Go 惯例、`go vet` 零告警、核心路径测试断言精确。这是一套有真实工程意识的模板库。

但审计揭示了一个**核心矛盾**：

> 项目通过 `goalcli score` 给自己打 **满分 10.0**，但该评分的 10 个维度**全部是文件存在性 / 字符串包含检查**，不执行任何运行时验证。"外壳严格，内核可绕过"。独立衡量真实质量约为 **7.1**；审计时**有 3 个测试正在失败**，修复对齐状态见第六节。

自评 10.0 与独立 7.1 之间 **2.9 分的差距**，正是这套庞大治理体系最需要关注的信号。

---

## 二、评分总表

| 维度 | 权重 | 得分 | 一句话结论 |
| --- | --- | --- | --- |
| 架构与代码质量 | 20% | **7.5** | API 设计规范，但 health.go 重复、releasemanifest 单文件 649 行 |
| 测试质量与覆盖率 | 25% | **7.5** | 公共层断言扎实（覆盖率 71.9%），但发布核心包零覆盖 + 3 个测试失败 |
| 文档与治理体系 | 25% | **7.0** | 内部一致性强，但文档:代码 ≈ 5:1，存在 AI 仪式化膨胀 |
| 安全 / 供应链 / CI / 发布 | 30% | **6.5** | 流程设计完整，但 Actions 未 pin SHA、score gate 可绕过 |
| **加权综合** | 100% | **7.1** | 优秀的参考实现，被形式主义的自评门禁高估 |

> 安全供应链权重最高（30%），因为本仓库的核心交付物之一就是"发布治理标准 + Evidence runtime"，其自身供应链与门禁的可信度是项目立身之本。

---

## 三、按优先级的修复清单

### 🔴 CRITICAL（审计时阻断发布，须立即修复）

| # | 问题 | 位置 | 修复建议 |
| --- | --- | --- | --- |
| C1 | 审计时 **3 个测试因依赖已被 git 删除的瞬态文件而失败**，CI 实际为 FAIL | `internal/tools/releasemanifest/main_test.go:462,532,650`（依赖 `.omc/state/agent-replay-*.jsonl`） | 测试应在 `t.TempDir()` 构造所需文件，禁止依赖工作区瞬态状态 |
| C2 | 审计时 **发布质量核心包零测试覆盖**（0%） | `internal/releasequality/score.go:28-80`（Compute/Verify/Marshal） | 为 Compute/Verify 各加单测，覆盖空输入与正常路径 |
| C3 | 审计时 **所有 GitHub Actions 未 pin commit SHA**，全用浮动 `@v4/@v5` tag，供应链攻击标准入口 | `ci.yml:15,18,23,63`、`release.yml:13,15,50`、`security.yml:14,16`、`integration.yml:14,16` | 改为 `actions/checkout@<40位SHA> # vX.Y.Z`，配合 dependabot 更新 |

### 🟠 HIGH（应在下个版本修复）

| # | 问题 | 位置 | 修复建议 |
| --- | --- | --- | --- |
| H1 | **score gate 全部是静态文本匹配，不执行运行时验证**，可通过在文件写入对应字符串骗取满分 | `internal/releasequality/score.go:33-42` | 对 security_gate/score_gate 增加执行验证维度，或在报告中标注语义边界 |
| H2 | 审计时 **secret scan 排除目录拼写错误**：写的是 `.omx`，实际目录是 `.omc`，导致 `.omc/` 未被排除 | `scripts/check_secrets.sh:25` | 改为 `--exclude-dir=.omc`，并审查 `.omc/` 是否含敏感数据 |
| H3 | 审计时 **govulncheck 使用浮动安装版本**，破坏构建可重现性 | `ci.yml:34`、`release.yml:23`、`security.yml:21` | 固定为 `govulncheck@v1.x.x`，与 golangci-lint 锁版本方式一致 |
| H4 | **CLI 入口零覆盖**（0%） | `cmd/goalcli/main.go:18-101` | 提取逻辑到可注入 `io.Writer` 的纯函数，table-driven 测试正常/错误路径 |
| H5 | **traceability-matrix 可追溯性停留在"文件指向"**，Evidence 列为占位符而非真实 CI run ID / commit | `.agent/traceability-matrix.md`（10 个 REQ） | 每个 REQ 补最近一次验证的 CI run ID 或 commit hash |
| H6 | **文档:代码比例严重失衡**（文档 7433 行 vs 非测试代码 1452 行 ≈ 5:1），.agent/ 35 个文件多为 AI 仪式化膨胀 | 全局 / `.agent/` | 精简 .agent/ 为 goal + traceability-matrix + decision-log 三份核心，其余归档 |

### 🟡 MEDIUM

| # | 问题 | 位置 |
| --- | --- | --- |
| M1 | health.go 7+ 处重复构造 `HealthStatus{}` 字面量（158 行） | `pkg/templatex/health.go:43-134`，提取 `makeStatus` 辅助函数 |
| M2 | releasemanifest 单文件 649 行职责过多（KISS 违规） | `internal/tools/releasemanifest/main.go`，按职责拆分子文件 |
| M3 | scorecard 10 个维度无一衡量真实代码质量（覆盖率/API 兼容/benchmark） | `docs/scorecard.md`，增加代码实质维度 |
| M4 | sanitize 全量替换为 `***`，无分级、无法区分空值与脱敏值、不处理结构化日志 | `internal/sanitize/sanitize.go` |
| M5 | risk-register 仅 4 条泛化条目，与 Full Goal Runtime 仪式规模不匹配 | `.agent/risk-register.md`，合并进 decision-log 或删除 |

### 🟢 LOW

- `SanitizedConfig` 类型冗余（YAGNI），全库除 `Sanitize()` 返回外无使用 — `pkg/templatex/config.go:16-20`
- `Version` 常量硬编码 `"v0.1.0"` 与实际发布版本脱节 — `pkg/templatex/version.go:4`
- `releasequality/score.go` 用相对路径探测文件（CWD 依赖），无文档说明
- testkit fixture 仅一个变体，复用价值有限
- .agent/ 模板文件与 docs/standard/ 内容重复，存在双重维护漂移风险

---

## 四、分维度详述

### 架构与代码质量 — 7.5
**优点**：错误类型设计完整（Kind/Op/Message/Cause/Retryable + errors.As 链，errors.go:24-43）；functional options 模式惯用（options.go）；并发安全到位（client.go:49-68 持锁最小化）；contracts 测试用反射比对 schema 与 Go 常量防协议漂移；internal 边界清晰。
**主要扣分**：health.go 重复代码、releasemanifest 巨型单文件、治理层认知负担偏重。

### 测试质量与覆盖率 — 7.5（实测 71.9%）
**优点**：断言精确到错误类型与 metrics 标签值，无"只跑不验证"；fuzz/property/golden/contracts 分层完整；testkit 克制（<50 行）；golden 无自动更新逻辑，强制显式更新。
**主要扣分**：发布门禁核心包 `internal/releasequality` 与 CLI 入口零覆盖；审计时 3 个测试因瞬态文件依赖持续失败——恰恰是最关键的发布评分逻辑缺乏保护，与"质量模板"定位矛盾。C1/C2 已在修复对齐中关闭；CLI 入口覆盖率仍需后续补齐。

### 文档与治理体系 — 7.0
**优点**：三篇 ADR 结构完整（背景-决策-影响-证据 + 拒绝的替代方案）；DoD 四层递进且含诚实的"无法完成"协议；安全策略具体可执行；抽查 15 个相对链接全部有效。
**判断**：结构价值真实（ADR/DoD/Evidence 协议/harness-gates 真正减少歧义），但**规模是 AI 仪式性膨胀**——.agent/ 中约 20 个文件（spec/plan/state-machine/object-model/各类 patch/agent-teams）描述了一个远比 1500 行代码复杂的"运行时"，属过度预制。

### 安全 / 供应链 / CI / 发布 — 6.5
**优点**：release-final-check 三道独立阀门（Makefile:128）；manifest evidence 链 SHA256 可重复校验；govulncheck 强制失败（非 soft-skip）；secret scan 覆盖主流 provider token；dependabot 双生态覆盖。
**判断**：**外壳严格，内核可绕过**。审计时的 Actions 未 pin SHA（C3）、govulncheck 浮动安装版本（H3）、score gate 纯文本匹配（H1）、secret scan 排除目录拼写 bug（H2）共同构成系统性形式主义风险，与整套 evidence 体系的防护意图直接矛盾。

---

## 五、结论

xlib-standard 是一个**优秀的 Go 参考实现，但被一套形式主义的自评门禁高估了**。

- **真正的资产**：`pkg/templatex` 的 API/错误/并发设计、contracts 反射对齐测试、ADR 与 DoD 的工程诚信。
- **真正的风险**：自评 10/10 来自可绕过的文件存在性检查；审计时 3 个测试失败；供应链入口（Actions SHA、govulncheck 版本）审计时未加固；治理文档膨胀到代码量的 5 倍。

**核心建议（按顺序）**：
1. C1（失败测试）、C2（核心包零覆盖）、C3（Actions pin SHA）已在 2026-06-02 修复对齐中关闭；后续发布前必须把这些项作为回归检查保留。
2. 让 score gate 名副其实：为关键维度加运行时验证，否则它只是"文件齐全勋章"。
3. 大幅精简 .agent/ 仪式工件，让治理重量与代码规模重新匹配，降低新人上手成本。

> 修复 CRITICAL + 关键 HIGH 后，本项目独立评分预计可达 **8.5 / 10**，届时自评 10.0 才真正具备公信力。

---

## 六、修复对齐状态（2026-06-02）

本节记录审计后的修复对齐事实，不重新计算独立综合评分；**7.1 / 10** 仍代表审计时快照。后续如需更新分数，应重新运行独立审计，而不是只按本表推导。

| 编号 | 状态 | 对齐证据 |
| --- | --- | --- |
| C1 | 已关闭 | `internal/tools/releasemanifest/main_test.go` 改为在临时 fixture 仓库内构造 `.omc/state/agent-replay-fixture.jsonl`，不再依赖工作区瞬态 Agent 运行态文件。 |
| C2 | 已关闭 | 新增 `internal/releasequality/score_test.go`，覆盖 `Compute`、`Verify` 和 `Marshal` 的正常与失败路径。 |
| C3 | 已关闭 | `.github/workflows/*.yml` 中的 `actions/checkout`、`actions/setup-go`、`actions/cache` 和 `actions/upload-artifact` 已固定为 40 位 commit SHA，并保留 tag 注释便于审计。 |
| H2 | 已关闭 | `scripts/check_secrets.sh` 同时排除 `.omc` 和 `.omx` 本地运行态目录，避免扫描 Agent 状态文件造成误报。 |
| H3 | 已关闭 | CI / release / security workflow 安装 `golang.org/x/vuln/cmd/govulncheck@v1.3.0`，不再使用 `@latest`。 |
| H1 | 部分对齐 | `docs/scorecard.md` 已补充 score 的语义边界；运行时验证维度仍需后续实现。 |
| H4-H6 与 Medium/Low | 未关闭 | 仍按原清单跟踪，未在本轮修复中变更 CLI 覆盖率、traceability evidence 或治理文档规模。 |

本轮修复后的本地验证证据：

- `go test ./...`
- `go vet ./...`
- `make lint`
- `make security`
- `make ci`

*本报告由 xlib-audit agent team（arch-reviewer / test-reviewer / docs-reviewer / secops-reviewer，均 Opus 模型）并行独立分析后综合生成。*
