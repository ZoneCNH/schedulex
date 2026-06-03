# Evidence 规则

> 源自 Goal 完整规则 v1.0 §12

## RULE-EVIDENCE-001：Evidence 是完成证明

Evidence 必须包含：

```text
evidence_id          # EVID-<task-id>-YYYYMMDD-NNN
related_goal
related_task
related_requirement
related_ac
command
output
artifact_path
timestamp
status
```

## RULE-EVIDENCE-002：Evidence 必须可复查

禁止：

```text
测试通过了
已完成
应该没问题
已修复
```

必须有：

```text
命令
输出
日志
报告
文件路径
PR 链接
CI 链接
```

## RULE-EVIDENCE-003：Evidence 必须进入 Traceability Matrix

```text
Requirement → AC → Task → Test → Evidence → Status
```
