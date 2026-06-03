# Truth-State 标准

Truth-State 是 first-PR/MVA 的治理事实分离标准，用于把“已登记”“已扫描”“可 dry-run”“已真实执行”“可作为 release evidence”“下游已采用”等状态拆成不同的本地事实文件。该标准只定义状态语义和最小验收关系，不生成 release manifest，不裁决 release-check，也不推进 migration-wave。

## 目的

`xlib-standard` 的治理结论必须来自可核验事实，不能把 registry、baseline scan、dry-run、artifact upload 或 downstream 声明混写为同一个 passed 结论。

first-PR/MVA 必须至少区分：

- 规则是否存在且被 Gate 引用。
- 命令是否只是 planned / dry-run，还是已有真实实现。
- Release required gate 是否有真实执行证据。
- Evidence artifact 是否可用、可校验、可追溯。
- Downstream 是否只是 registered，还是已有 proof-based adoption。

## MVA 状态文件

| 文件 | 职责 | 不得表示 |
| --- | --- | --- |
| `.agent/truth-state.yaml` | 汇总治理状态分类、允许状态值和 first-PR/MVA scope。 | 不得声明 release 已通过或 downstream 已 adopted。 |
| `.agent/command-implementation-status.yaml` | 区分 command registered、planned、dry-run verify 和 implemented。 | 不得把 planned command 记为真实 gate success。 |
| `.agent/release-required-gates.yaml` | 列出 release required gate 与当前证据等级。 | 不得生成或替代 `release/manifest/latest.json`。 |
| `.agent/evidence-usability.yaml` | 区分 artifact exists、checksum verified、replayable、usable。 | 不得把 artifact upload 等同 release usable。 |
| `.agent/downstream-adoption-status.yaml` | 区分 downstream registered、baseline scanned、patch planned、proof adopted。 | 不得把 registry 存在等同 adopted。 |

这些文件是源码中的治理事实 contract。它们可以被 `governance-check` 或后续 goalcli contract 检查读取，但 first-PR/MVA 不要求改变 release runtime、release manifest schema 或 migration-wave 运行时。

## 状态语义

| 状态 | 含义 | 可作为 passed gate evidence |
| --- | --- | --- |
| `registered` | 已在 registry、文档或 baseline 中声明。 | 否。 |
| `baseline_scanned` | 已完成静态或 baseline 扫描。 | 否，除非 gate 明确只要求 scan。 |
| `planned` | 有计划、命令入口或 contract，但未实现真实行为。 | 否。 |
| `dry_run_ready` | 可执行本地 dry-run 或 contract 检查。 | 仅可作为 dry-run gate evidence。 |
| `implemented` | 已有真实本地实现，可在要求 context 下执行。 | 取决于命令输出。 |
| `executed` | 本次 Evidence 中有命令输出和退出码。 | 是，但必须保留命令证据。 |
| `artifact_exists` | artifact 文件或 CI 上传存在。 | 否。 |
| `checksum_verified` | artifact 校验和与记录一致。 | 只证明完整性，不证明 usable。 |
| `usable` | artifact 可被目标流程消费，且有 replay 或 contract 证据。 | 是，限对应 gate。 |
| `adopted` | downstream 有 proof-based 使用证据。 | 是，限 downstream adoption gate。 |

## Gate 关系

`governance-check` 可以使用 Truth-State 文件证明状态分类存在且未把弱事实升级为强结论。Release 相关 gate 必须继续依赖真实 gate 输出、checksum、manifest 和 Evidence contract：

```text
registered != adopted
baseline_scanned != implemented
dry_run_ready != executed
artifact_exists != usable
CHECK_STATUS=passed != release-ready evidence
```

first-PR/MVA 的通过条件是：状态文件存在、语义清晰、docs-check 可发现本标准，并且 `GOWORK=off make governance-check` 不因缺少该标准失败。

## DONE with evidence 要求

涉及 Truth-State 的完成声明必须记录：

- 修改的 `.agent/*status*.yaml` 或 `truth-state.yaml` 文件。
- 是否触达 release manifest 或 release-check；first-PR/MVA 应明确写为未触达。
- 至少一个本地命令证据，例如 `GOWORK=off make governance-check`。
- 已知缺口：后续 gate-result runtime、release manifest cutover、artifact provenance、downstream migration-wave 若未实现必须继续列为 follow-up。
