# Release 规则

> 源自 Goal 完整规则 v1.0 §16

## RULE-RELEASE-001：Release 必须有 Release Manifest

Release Manifest 必须包含：

```text
version
date
goal
commit
tag
included issues
included PRs
changes
evidence summary
test summary
compatibility
migration notes
risks
rollback plan
known issues
retrospective
```

## RULE-RELEASE-002：Release 前必须全部验证

必须通过：

```text
make release-check
make evidence-check
make ci
```

## RULE-RELEASE-003：Release 必须可回滚

必须有：

```text
rollback command
rollback condition
last known good version
affected components
risk note
```

## Release 验收清单

```text
[ ] Issues 已关闭
[ ] PRs 已合并
[ ] Evidence 已归档
[ ] CHANGELOG 已更新
[ ] VERSION 已更新
[ ] Release Manifest 已生成
[ ] Rollback Plan 已记录
```
