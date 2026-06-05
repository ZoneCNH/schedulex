# goalcli evidence ledger

`.agent/evidence/ledger.jsonl` 是 goalcli v0.1.0 的 source evidence ledger。`release/evidence/goalcli/` 可以保存 generated evidence packs，但这些 release artifacts 不是 source of truth，也不能作为 canonical ledger input。

G12-G16 commands 是由 `goalcli` 和 Harness Runtime 背书的 goalcli MVA-blocking checks。fresh ledger-backed evidence 只有在 `goal-runtime-final` 调和同一 `GOAL_ID` 下的 `goal-acceptance`、`goal-delivery`、`goal-handover`、`goal-downstream-adoption` 和 `goal-certify` 后，才能报告 `mva_status: complete`。
