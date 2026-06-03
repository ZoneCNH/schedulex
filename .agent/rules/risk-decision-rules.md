# Risk / Decision / Rollback 规则

> 源自 Goal 完整规则 v1.0 §20

## RULE-RISK-001：P0/P1 风险必须登记

Risk Register 字段：

```text
risk_id
description
probability
impact
severity
affected_objects
mitigation
fallback
owner
status
```

## RULE-DECISION-001：关键选择必须记录

Decision Log 字段：

```text
decision_id          # DEC-YYYYMMDD-NNN
context
options
selected_option
reason
tradeoff
affected_objects
rollback_condition
```

## RULE-ROLLBACK-001：高风险变更必须可回滚

涉及以下内容必须写 rollback：

```text
CI
release
storage
config
public API
security
automation
harness
rules
```
