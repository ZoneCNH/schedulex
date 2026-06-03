# Evidence Template

```text
DONE with evidence:
- scope: <task|issue|goal|release>
- summary: <one sentence>
- gates:
  - GOWORK=off make ci: <passed|failed|blocked> <evidence>
  - GOWORK=off make integration: <passed|failed|blocked> <evidence>
  - CHECK_STATUS=passed GOWORK=off make evidence: <passed|failed|blocked> <artifact>
  - RELEASE_EVIDENCE_REQUIRE_PASSED=1 GOWORK=off make release-evidence-check: <passed|failed|blocked> <evidence>
- artifacts:
  - release/manifest/latest.json: <generated|not generated and why>
- changed files:
  - <path>
- known gaps:
  - <none or explicit blocker>
```

Required gate 失败时不得把 scope 标记为 release complete。
