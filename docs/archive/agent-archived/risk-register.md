# Risk Register

| Risk | Impact | Mitigation | Owner |
| --- | --- | --- | --- |
| Legacy naming remains in non-migration docs | REQ-002 failure | Scan and classify allowed migration hits | docs/runtime worker |
| Docs claim executable gate before implementation exists | False Evidence | Mark executable-gate gaps with owner until worker-3/4 evidence lands | leader + gate workers |
| x.go boundary ambiguity | Coupling regression | Maintain x.go integration boundary and boundary gate | docs + gate workers |
| Secret path leaks into artifacts | Security failure | Treat `/home/k8s/secrets/env/*` as caller-only path; scan Evidence | security worker |
