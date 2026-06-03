# Self-improving 规则

> 源自 Goal 完整规则 v1.0 §17

## RULE-RETRO-001：每个 Goal 必须有 Retrospective

必须回答：

```text
什么有效
什么失败
根因是什么
哪个 Gate 缺失
哪个 Rule 缺失
哪个 Prompt 需要修复
哪个 CI 需要新增
下轮如何自动避免
```

## RULE-RETRO-002：必须生成 Patch

至少生成：

```text
Prompt Patch          PATCH-PROMPT-YYYYMMDD-NNN
Harness Patch         PATCH-HARNESS-YYYYMMDD-NNN
Rule Patch            PATCH-RULE-YYYYMMDD-NNN
CI Gate Suggestion
New Issue Candidates
```

## RULE-RETRO-003：重复问题必须升级为规则

如果同类问题出现两次，必须：

```text
加入 rule
加入 harness gate
加入 CI check
加入 template
```

## Self-improving 验收清单

```text
[ ] Retrospective 已生成
[ ] Prompt Patch 已生成
[ ] Harness Patch 已生成
[ ] Rule Patch 已生成
[ ] New Issue Candidates 已生成
```
