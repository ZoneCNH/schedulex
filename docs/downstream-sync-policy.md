# 下游同步策略

本文定义 `xlib-standard` 变更后如何同步到下游基础库。`xlib-standard` 是 Standard Source、Go Reference Template、Generator、Harness 和 Evidence Runtime 的唯一标准源；下游只消费这些标准和模板结果，不反向决定标准内容。

旧 `baselib-template` 和 `foundationx` 名称只保留在迁移文档语境中。当前持久同步目标使用 `kernel` 与 L1/L2 基础库命名；`corekit` 仅作为中性路径 smoke/integration 验证目标，`xlib-standard` 是标准源而不是下游同步目标。

## 角色

| 角色 | 当前名称 | 同步关系 |
| --- | --- | --- |
| Standard Source | `xlib-standard` | 定义标准、边界、gate、Evidence 和下游同步规则 |
| Go Reference Template | `xlib-standard` | 提供可渲染模板和参考实现 |
| Generator | `scripts/render_template.sh` / `cmd/goalcli integration` | 把模板渲染为具体基础库 |
| L0 代表下游 | `kernel` | 第一优先级同步目标，验证最小基础库形态 |
| L1 基础库 | `configx`、`observex`、`testkitx` | 继承 L0 标准并提供基础能力 |
| L2 基础库 | `postgresx`、`redisx`、`kafkax`、`taosx`、`ossx`、`clickhousex` | 在 L1 能力上提供具体基础设施适配 |
| 组合层 | `x.go` | 仅作为消费方组合基础库，不反向影响 `xlib-standard` |

## 当前采纳状态口径

`docs/downstream-matrix.md` 只声明目标库和同步范围，不等于下游已采纳证据。当前 proof-based adoption 以 [.agent/downstream-adoption-status.yaml](../.agent/downstream-adoption-status.yaml) 和 [.agent/truth-state.yaml](../.agent/truth-state.yaml) 为准；`registered`、`baseline_scanned`、`patch-only` 或 dry-run 结果不得升级为 `adopted`。没有当前 downstream 命令输出时，完成声明只能记录 blocked、known gap 或待同步 owner。

字段语义必须按下表解释，避免把登记态或 dry-run 状态误写成采纳证据：

| 字段 | 语义 | 禁止升级 |
| --- | --- | --- |
| `mode` | 当前执行或计划模式，例如 `patch-only` | 不得作为 proof-based adoption |
| `status` | 当前 worker workspace 中的可用性或 gap 状态 | gap 状态不得作为 passed evidence |
| `registry_state` | 标准源声明的目标登记态 | `standard_target_declared` 不等于 downstream implemented |
| `adoption_status` | downstream 采纳结论 | 只有当前 downstream 命令输出能支持 `adopted` |
| `evidence_state` | 证据覆盖状态 | `not_run`、`registry_gap_only` 或 dry-run 不得作为 release gate passed evidence |

`.agent/truth-state.yaml` 的 forbidden upgrades 是该口径的单一约束：`registered != adopted`、`baseline_scanned != adopted/implemented`、`patch_only != proof_based_adoption`、`not_run != passed`。

`Context Runtime v4.0 profile` 是 context profile / registry bridge 的同步条目名称，不改写 Full Goal Runtime v3.1 的 `.agent/` 工件版本，也不替代 [docs/goal.md](goal/goal.md) v2.9.3 Complete 的治理主基线。

## 变更到同步动作映射

| `xlib-standard` 变更类型 | 必需同步动作 | Evidence |
| --- | --- | --- |
| `docs/standard/**` 标准文本、仓库角色、分层或模块边界变更 | 更新下游 README、标准引用、DoD 和边界说明；检查旧名是否只在迁移文档语境出现 | `GOWORK=off make docs-check`，必要时附 `release/standard-impact/latest.md` |
| `contracts/**`、metrics、health JSON 或 config schema 变更 | 通知所有受影响基础库更新 contract、测试和示例；breaking change 必须进入 release notes | contracts gate、下游 contract 测试、`downstream_release_decision: required` 结论 |
| `scripts/render_template.sh`、generator 占位符或包目录规则变更 | 重新渲染 `kernel`，并运行 `corekit` 中性路径 smoke；确认 module path、package name、README、docs 和 contracts 无旧模板残留，必要时扩展到 L1/L2 采用目标 | `GOWORK=off make integration` |
| Harness gate、Makefile、CI 或 `.agent/harness.yaml` 变更 | 同步下游 gate 文档和 CI 入口；强制 gate 不得在下游降级为可选 | `GOWORK=off make release-check` 或对应下游 gate 输出 |
| Evidence protocol、release manifest 字段或 artifact 规则变更 | 更新下游 Evidence 生成、校验和发布模板；manifest 字段变化必须标记同步需求 | manifest 校验、checksum、CI artifact |
| Context Runtime v4.0 profile、registry bridge、`.agent/context/*` 或 `templates/context-consumer/*` 变更 | 同步下游 context profile 入口、legacy alias、registry 引用和运行时证据；profile wrapper/registry bridge 按当前 gate 事实同步，物理 `.agent/context/*` 或 `templates/context-consumer/*` 未落地前不得宣称下游可消费这些文件 | `context_runtime` / `governance_registry` / `repository_rules` / `downstream_context` taxonomy，`governance_runtime` manifest Evidence |
| 依赖或安全策略变更 | 判断是否影响所有基础库；安全变更默认触发下游同步 | `govulncheck`、secret scan、依赖清单 |
| 命名、仓库角色或默认下游变更 | 当前主叙事必须使用 `xlib-standard` 和 `kernel`；`corekit` 只用于中性路径 smoke/integration 语境；旧名只能保留在迁移文档语境 | `docs-check` 命名残留断言 |

`downstream_release_decision` 的 allowed values 只能是 `required` 或 `not_required`。当标准影响报告判定需要同步下游时使用 `required`，否则使用 `not_required`。

`repository_rules_release_decision` 的 allowed values 只能是 `audit_required` 或 `not_required`。当仓库规则变更需要对下游执行审计时使用 `audit_required`，否则使用 `not_required`。

## `kernel` 同步规则

`kernel` 是 L0 代表下游，也是默认生成验证目标。任何触达模板、generator、公共 API、contracts、Harness gate、Evidence protocol、release manifest 或仓库角色命名的变更，都必须判断是否需要同步到 `kernel`。

必须同步到 `kernel` 的情形：

- 模板渲染输出、package 目录、module path 或 README/docs 占位符发生变化。
- Standard、DoD、module boundary、Harness gate 或 Evidence 要求发生变化。
- public API、config、error、health、metrics、contracts 或 release manifest schema 发生变化。
- `release/standard-impact/latest.md` 或同类报告标记 `downstream_sync_required=true` 且 `downstream_release_decision=required`，release manifest 对应字段也必须一致。

`downstream_release_decision` 的 allowed values 只能是 `required` 或 `not_required`：需要同步默认下游时使用 `required`，确认无需同步或影响已被排除时使用 `not_required`。

同步完成前，变更说明必须记录 `kernel` 的状态：已同步、无需同步并说明原因，或 blocked 并列出阻塞条件。

## L1 基础库同步规则

L1 基础库包括 `configx`、`observex` 和 `testkitx`。它们继承 L0 标准，并为其他基础库提供基础配置、观测和测试能力。

必须同步到 L1 的情形：

- Standard、Harness gate、Evidence protocol 或 security policy 发生跨库规则变化。
- `config`、metrics、health、testkit fixture 或错误分类规则发生变化。
- `kernel` 同步后发现 L1 文档、contracts 或 gate 与新标准不一致。

L1 同步不得引入 `x.go` 业务模型、生产密钥路径读取或应用 wiring。L1 可以记录 `x.go` 作为消费方，但不能依赖 `x.go`。

## L2 基础库同步规则

L2 基础库包括 `postgresx`、`redisx`、`kafkax`、`taosx`、`ossx` 和 `clickhousex`。L2 只在变更影响具体适配、contracts、release Evidence、security policy 或共享 gate 时同步。

L2 同步优先级低于 `kernel` 和 L1；若 `release/standard-impact/latest.md` 未标记 L2 影响，可以只在 release note 中记录无需同步的原因。

## `x.go` 消费方规则

x.go 仅作为基础库消费方和应用组合层。它可以根据自身需要拉取 `kernel`、L1 或 L2 基础库的新版本，但不得反向要求 `xlib-standard` 引入业务模型、业务 repository、业务消息 schema、生产密钥读取或应用 wiring。

当 `x.go` 暴露出基础库标准缺口时，应先在 `xlib-standard` 形成标准变更、ADR 或 Issue，再由标准源决定是否触发下游同步。不得把 `x.go` 的临时实现直接复制回 `xlib-standard`。

## PR 与发布要求

- 触达标准、模板、generator、Harness、Evidence、contracts、命名或下游矩阵的 PR，必须说明是否触发 `downstream_release_decision: required`，并记录 release manifest 的 `downstream_sync_required` / `downstream_release_decision` 结论。
- 触发同步时，PR 或 release Evidence 必须列出 `kernel`、L1、L2 和 `x.go` 的影响结论。
- 未完成同步时，不得在完成声明中写 “无需下游动作”；必须写明 blocked 原因和后续 owner。
- `GOWORK=off make docs-check` 必须校验本文件存在、关键角色命名和旧名限制。
