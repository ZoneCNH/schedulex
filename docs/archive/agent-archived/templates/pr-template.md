# PR 模板

> 源自 RULE-PR-001 ~ RULE-PR-004

## 关联

- Goal: GOAL-YYYYMMDD-NNN
- Issues: #NNN
- Requirements: REQ-xxx-NNN

## 变更说明

-

## Traceability Matrix

| Requirement | AC | Task | Test | Evidence | Status |
|-------------|-----|------|------|----------|--------|
| REQ-xxx-NNN | AC-xxx-NNN | TASK-xxx-NNN | TEST-xxx-NNN | EVID-xxx | |

## 验证

- [ ] `GOWORK=off make ci`
- [ ] `GOWORK=off make integration`
- [ ] `CHECK_STATUS=passed GOWORK=off make evidence`
- [ ] `GOWORK=off make release-check`

## Evidence

- Evidence ID:
- CI artifact:
- Known gaps:

## 边界检查

- [ ] 未引入 `x.go` 依赖
- [ ] 未引入业务模型或业务流程
- [ ] 未提交真实凭据或生产连接串

## 风险

- 级别:
- 回滚方案:
