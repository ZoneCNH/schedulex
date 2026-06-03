# 状态机

```text
intake -> scope_lock -> plan -> implement -> verify -> review -> release -> retrospective -> complete
                         |          |          |          |             |
                         v          v          v          v             v
                      blocked <--- fix <--- changes_requested <--- rollback
```

## 状态

- `intake`: goal/context/task 已加载；owner 已识别。
- `scope_lock`: worker scope 和 forbidden files 已记录。
- `plan`: required artifacts、AC、risk 和 verification commands 已映射。
- `implement`: 已编辑限定 scope 内文件。
- `verify`: 已运行 tests、docs-check、boundary/contracts/integration/release/score checks，或已记录 gaps。
- `review`: reviewer 验证 Evidence 和 scope compliance。
- `release`: manifest、checksum、version 和 final gate 已记录。
- `retrospective`: defects 回流到 prompt/harness/rule patches。
- `complete`: DONE with evidence，且没有 open blocker。
- `blocked`: owner/action 已记录；不得静默部分完成。
- `rollback`: 按 rollback protocol 执行 revert 或 mitigation path。

## 转换规则

- `implement` 不能在 scope lock 之前开始。
- `complete` 要求所有 REQ 在 traceability matrix 中关闭。
- `release` 要求 `GOWORK=off make release-final-check` 和 score gate；如果仍缺失的 executable gate 由其他 worker 负责，本 slice 必须记录 gap。
