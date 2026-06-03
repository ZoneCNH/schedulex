# PR 说明

## 影响范围

- 模板行为：
- 生成库影响：
- 标准或 contract 影响：

## 验证

- [ ] `GOWORK=off make ci`
- [ ] `GOWORK=off make integration`
- [ ] `CHECK_STATUS=passed GOWORK=off make evidence`
- [ ] `RELEASE_EVIDENCE_REQUIRE_PASSED=1 GOWORK=off make release-evidence-check`
- [ ] `GOWORK=off make release-check`

## Evidence

- `release/manifest/latest.json`：
- CI artifact：
- known gaps：

## 边界检查

- [ ] 未引入 `x.go` 依赖。
- [ ] 未引入业务模型或业务流程。
- [ ] 未提交真实凭据或生产连接串。
- [ ] 未提交 `release/manifest/latest.json`。

## 关联

- Issue：
- ADR：
