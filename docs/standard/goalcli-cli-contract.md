# goalcli CLI Contract

`goalcli` 是 xlib-standard v2.9.3 的 machine-verifiable gate surface。除明确委托给既有脚本的命令外，所有命令必须输出包含 `command`、`status`、可选 `details` 和可选 `gaps` 的 JSON report，并符合 `contracts/goalcli-report.schema.json`。

该 CLI 只执行本地非破坏性 contract checks：不读取真实 secrets，不连接外部生产系统，也不修改 downstream 仓库。涉及 GitHub、release、runtime 或 downstream 状态的命令，在当前版本中以本地 contract、manifest 和 dry-run 证据表达。

## 退出码语义

- `status=passed` 必须返回 `0`。
- `status=failed`、`status=planned` 和 `status=gap` 必须返回 `1`。
- 未知命令、非法参数或非法 context 必须返回 `2`。
- downstream 仓库不存在时必须报告 `status=gap` 且返回非 0；该结果不能作为 release gate 成功证据。
- 带 `--verify` 或 `--strict` 的命令遇到 `planned`/`gap` 时必须阻断 gate。

## P0 commands

- `version`
- `doctor`
- `minimal-kernel`
- `main-guard --context local_write|local_readonly|ci_pull_request|ci_main_verify|release_verify`
- `worktree-guard --context local_write|local_readonly|ci_pull_request|ci_main_verify|release_verify`
- `worktree-check --context local_write|local_readonly|ci_pull_request|ci_main_verify|release_verify`
- `context-check`
- `spec-check`
- `design-check`
- `task-check`
- `pr-check`
- `evidence-check`
- `done-assertion`
- `cli-contract`
- `issue-registry`
- `command-registry`
- `makefile-baseline`
- `context-profile`
- `context-profile-check`
- `context-schema-check`
- `context-fast-check`
- `context-standard-check`
- `context-full-check`
- `boundary`
- `contracts`
- `dependency-check`
- `docs-check`
- `evidence`
- `manifest`
- `integration`
- `rules-verify`
- `release-evidence-check`
- `release-evidence-checksum-check`
- `release-evidence-hash`
- `release-final-check`
- `render-check`
- `rules-consistency-check`
- `score`
- `secrets`
- `security`
- `standard-impact-check`
- `debt`
- `architecture`
- `domain`
- `docs-drift`
- `dependency-debt`
- `security-debt`
- `testing-debt`
- `implementation-debt`
- `downstream-debt`
- `debt-evidence`
- `debt-evidence-checksum-check`
- `debt-evidence-hash`

## P1 commands

- `agent-team-contract`
- `scope-lock`
- `pr-template`
- `acceptance-matrix`
- `runtime-health`
- `goal-runtime`
- `goal-acceptance --goal-id <GOAL-ID> --mode FULL --json`
- `goal-delivery --goal-id <GOAL-ID> --mode FULL --json`
- `goal-handover --goal-id <GOAL-ID> --mode FULL --json`
- `goal-downstream-adoption --goal-id <GOAL-ID> --mode FULL --json`
- `goal-certify --goal-id <GOAL-ID> --mode FULL --json`
- `goal-runtime-final --goal-id <GOAL-ID> --mode FULL --json`
- `naming`
- `upgrade-standard --dry-run --repo <path>`
- `conformance-profile`
- `downstream-registry`
- `self-healing-skeleton`
- `policy-schema`
- `github-settings --verify`
- `github-governance`
- `governance-fixture-test`
- `toolchain`
- `evidence-artifacts`
- `supply-chain`
- `autoresearch`
- `changelog`

## P2 commands

- `install-runtime --dry-run`
- `upgrade-runtime --dry-run`
- `release-ready`
- `evidence-replay`
- `attest-conformance --profile standard-source|l0-kernel`
- `pack-standard`
- `pack-gate`
- `pack-evidence`
- `runtime-file-ownership`
- `downstream-baseline --repo kernel/configx --mode patch-only`
- `downstream-adoption --repo kernel/configx --mode patch-only`
- `execution-context`

## Goalcli MVA commands

Goalcli MVA commands（`goal-acceptance`、`goal-delivery`、`goal-handover`、`goal-downstream-adoption`、`goal-certify` 和 `goal-runtime-final`）用于证明 `goalcli v0.1.0` 的 G12-G16 evidence contracts，并且只在 goalcli MVA evidence scope 内是 blocking gate。它们不会创建第二套并列 goalcli 执行面，也不会修改 downstream 仓库。source authority 是 `cmd/goalcli` 执行面和 `.agent/evidence/ledger.jsonl`；`release/evidence/goalcli/` 下的 generated packs 只是派生产物。root goalcli plan 仍是 roadmap authority，完成证据来自同一 `GOAL_ID` 下已调和的 source ledger 和 final report。

## Debt governance commands

Debt governance commands are P0 release-blocking gates. `goalcli debt` runs the full debt scanner, while `architecture`, `domain`, `docs-drift`, `dependency-debt`, `security-debt`, `testing-debt`, `implementation-debt`, and `downstream-debt` run focused slices of the same policy. The scanner reuses existing local scripts for boundary, docs, dependency diff, and secret checks; scanner failures return non-zero and cannot be treated as passed evidence.

`goalcli debt-evidence` writes generated evidence to `release/debt/latest.json`, `release/debt/latest.md`, and `release/debt/latest.json.sha256`. These latest evidence files are reproducible release artifacts and are intentionally ignored by git. P0 debt rules are not exceptable: policy files under `.agent/debt/` must not introduce P0 exception markers, and release verification must fail if debt status is not `passed`.

## Goalcli v0.1.0 MVA runtime commands

goalcli MVA commands 是由 `.agent/harness.yaml` 背书的本地 `cmd/goalcli` evidence commands；它们把 command、Makefile 和 harness coverage 变成可执行契约。只有当同一 `GOAL_ID` 的 G12-G16 检查都已写入 source ledger 并由 final rollup 调和后，才能证明 goalcli v0.1.0 MVA completion。

- `goal-acceptance` 校验 `G12_ACCEPTANCE`。
- `goal-delivery` 校验 `G13_DELIVERY`。
- `goal-handover` 校验 `G14_HANDOVER`。
- `goal-downstream-adoption` 校验 `G15_DOWNSTREAM_ADOPTION`。
- `goal-certify` 校验 `G16_CERTIFY` 并保留 source-ledger evidence claim。
- `goal-runtime-final` 校验五个 prerequisite local gates 的 `G12_G16_FINAL` rollup。
- 直接运行 `goalcli <command> --json` 是只读检查，不会写 evidence。
- 只有显式传入 `--write-evidence` 时，命令才会写入 `.agent/evidence/ledger.jsonl`。
- `goal-runtime-final --write-evidence` 只有在 source ledger 已存在同一 `GOAL_ID` 的五个 prerequisite entries 后，才会写入 generated evidence pack。

## 执行约束

- `--dry-run` 只能执行本地 contract、manifest 或 patch planning 检查。
- `--verify` 或 `--strict` 不能把 `planned`/`gap` 当作成功证据。
- `--context` 仅允许 `local_write`、`local_readonly`、`ci_pull_request`、`ci_main_verify` 和 `release_verify`。
- `--repo` 指向的 downstream 仓库不存在时，命令必须返回 `gap`，且不得自动创建目录。
- 新增命令时必须同步 `run` dispatch、`plannedCommandFiles`、Makefile gate、CLI contract 和测试表。

## 语义校验

`issue-registry` 不只是文件存在检查。它必须校验 `.agent/issue-registry.yaml` 中每个条目都具备非空 `title`、`command` 和 `evidence`，`status` 必须为 `implemented`，ID 必须匹配 `P0|P1|P2|CTX-###`、全局唯一，并且每个前缀从 `001` 连续编号。`context-profile-check` 复用该 registry 语义，不能用空文件或非连续 ID 作为通过证据。

planned command 的 dry-run 也必须读取对应文件并检查语义 marker。当前强制 marker 包括：`agent-team-contract` 的 `schema_version:`、`roles:`、`rule:`；`acceptance-matrix` 的 `schema_version:`、`acceptance:`；`runtime-health` 的 `schema_version:`、`checks:`、`toolchain`；goalcli MVA 命令在 `.agent/harness.yaml` 中的 `goalcli_mva_gates:`、对应 `G12_ACCEPTANCE`/`G13_DELIVERY`/`G14_HANDOVER`/`G15_DOWNSTREAM_ADOPTION`/`G16_CERTIFY`/`G12_G16_FINAL` 与命令名；`execution-context` 的 `schema_version:`、`contexts:`、`local_write`、`ci_pull_request`、`release_verify`。这些命令不得退化为单纯路径存在检查。

## Context Runtime v4 命令

Context Runtime v4.0 新增可叠加的 profile baseline，但不替换现有 P0/P1/P2 command registry。以下命令受 registry 治理，必须同时保留在 `run` dispatch、`.agent/command-registry.yaml`、`.agent/issue-registry.yaml`、Makefile targets 和本 contract 中：

- `context-profile --profile lite|standard|full|release`
- `context-profile-check`
- `context-schema-check`
- `context-lite`
- `context-standard`
- `context-full`
- `context-release`
- `context-fast-check`
- `context-standard-check`
- `context-full-check`

`context-release` 不得调用 `release-check` 或 `release-final-check`；唯一允许方向是 `release-final-check` 在严格 release evidence check 前委托执行 `context-release`。Legacy aliases（`context-fast-check`、`context-standard-check`、`context-full-check`）必须保持可用。除非仓库内实际存在 `.agent/context` 文件，否则该基线不得宣称已落地这些文件。

## Debt command surface

`goalcli debt` supports `--section`, `--mode`, `--min-score`, and `--output json|markdown`. Aliases `architecture`, `domain`, `docs-drift`, `dependency-debt`, `testing-debt`, `implementation-debt`, and `security-debt` select the corresponding debt sections.
