# 复盘和补丁协议

复盘的目标是把失败转化为可执行补丁，避免同类问题重复出现。

## 触发条件

以下情况必须复盘：

- Required gate 失败。
- release Evidence 不一致。
- generator 生成结果不完整。
- 发现边界违规或 secret 风险。
- Review 发现标准文档和实现不一致。

## 输出格式

```text
Retrospective:
- failure:
- root cause:
- detection:
- fix:
- remaining risk:

Prompt Patch:
- <agent instruction update>

Harness Patch:
- <gate/script/check update>

Rule Patch:
- <standard/doc update>

CI Gate Suggestion:
- <workflow or make target change>
```

## Patch 分类

- Prompt Patch：修正 Agent 执行提示或检查清单。
- Harness Patch：新增或收紧脚本、Makefile target、manifest 校验。
- Rule Patch：更新 `docs/standard/`、contracts 或 ADR。
- CI Gate Suggestion：把本地发现转成 CI 可执行 gate。

## 收敛规则

- 只记录可执行补丁，不写泛泛建议。
- 已修复项必须带验证命令。
- 未修复项必须有 owner、阻塞原因或后续 Issue。
