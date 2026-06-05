# Object Model

| Object | Required fields | Evidence link |
| --- | --- | --- |
| Goal | id, target outcome, non-goals, final gates | docs/goal.md, .agent/goal.md |
| Spec | behavior, constraints, compatibility | docs/spec.md, .agent/spec.md |
| Requirement | REQ id, owner, status, AC ids | .agent/traceability-matrix.md |
| AC | observable condition, proof command/artifact | docs/goal.md |
| Design | boundary, interface, dependency direction | docs/design.md, .agent/design.md |
| ADR | decision, context, consequences | docs/adr/*.md |
| Plan | phases, sequencing, risk gates | .agent/plan.md |
| Task | id, owner/scope, lifecycle status | .agent/tasks.md, OMX task state |
| Test | command, expected result, owner | .agent/harness.yaml |
| Evidence | artifact, command, checksum, actor, timestamp | .agent/evidence-protocol.md |
| Risk | risk, impact, mitigation, owner | .agent/risk-register.md |
| Decision | decision, reason, rejected alternative | .agent/decision-log.md |
| Review | findings, verdict, required fixes | .agent/review-template.md |
| Release | version, gates, manifest, checksum | .agent/release-template.md |
| Retrospective | incident/learning, patch candidates | .agent/retrospective-template.md |
| Patch | prompt/harness/rule patch, trigger, test | .agent/prompt-patches.md, .agent/harness-patches.md, .agent/rule-patches.md |
