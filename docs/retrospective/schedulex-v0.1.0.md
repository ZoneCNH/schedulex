# schedulex v0.1.0 Retrospective

Prompt Patch: 后续目标应先列出必须删除的模板包和必须保留的治理文件，减少身份迁移歧义。

Harness Patch: schedulex gate 固定为 `go test ./...`、`make boundary`、`make schedulex-checks`、`make release-preflight VERSION=v0.1.0`。

Rule Patch: L1 调度器不得引入真实外部锁后端；只暴露接口和本地合同测试。
