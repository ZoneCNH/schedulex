# Debt Governance

`goalcli debt` is the release-facing technical debt scanner. It reads `.agent/debt/rules.yaml`, `.agent/debt/rule-registry.yaml`, `.agent/debt/exceptions.yaml`, and `.agent/debt/dependency-purpose.yaml`, then emits deterministic JSON or Markdown evidence.

Required gates:

- `make architecture` and `make domain` run enforce-mode P0 scans.
- `make docs-drift`, `make dependency-debt`, `make testing-debt`, `make implementation-debt`, and `make security-debt` cover non-architecture sections with warn/observe defaults.
- `make debt` enforces all sections with `--min-score 9.8`.
- `make debt-evidence` generates `release/debt/latest.json`, `release/debt/latest.md`, and `release/debt/latest.json.sha256`.

Release manifests include `debt` evidence with policy/report digests, score, status, active profile, and per-section P0/P1/P2 counts. Release verification rejects missing debt evidence, non-passing status, score below 9.8, or any P0 finding.
