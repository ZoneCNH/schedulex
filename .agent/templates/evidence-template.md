# Evidence 模板

> 源自 RULE-EVIDENCE-001 ~ RULE-EVIDENCE-003

## 基本信息

- Evidence ID: EVID-<task-id>-YYYYMMDD-NNN
- Goal: GOAL-YYYYMMDD-NNN
- Task: TASK-xxx-NNN
- Requirement: REQ-xxx-NNN
- AC: AC-xxx-NNN
- Timestamp:
- Status: passed | failed

## 执行记录

### 命令

```bash
# 执行的命令
```

### 输出

```text
# 命令输出
```

## Artifact

- 文件路径:
- CI 链接:
- PR 链接:

## Traceability

```text
Requirement → AC → Task → Test → Evidence → Status
REQ-xxx    → AC-xxx → TASK-xxx → TEST-xxx → EVID-xxx → passed
```

## 安全过滤确认

- [ ] 不包含 token / password / secret / private key
- [ ] 不包含 access key / cookie / authorization header
