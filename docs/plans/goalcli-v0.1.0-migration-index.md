# schedulex v0.1.0 migration index

Use this index to avoid false authority or completion claims while migrating from the root worktree proposal to canonical docs.

## Canonical paths

- Standard: `docs/standard/schedulex-runtime.md`
- ADR: `docs/adr/ADR-20260603-001-schedulex-runtime.md`
- Roadmap: `docs/plans/schedulex-v0.1.0-roadmap.md`
- Command registry: `.agent/command-registry.yaml`
- Harness control plane: `.agent/harness.yaml`
- Runtime registry: `.agent/registry/runtime.yaml`
- Source evidence ledger: `.agent/evidence/ledger.jsonl`

## Evidence split

`.agent/evidence/ledger.jsonl` is the canonical source ledger. `release/evidence/schedulex/` is reserved for generated evidence packs and must not be committed or treated as primary source evidence.
