# 测试策略

必须运行 `GOWORK=off go test ./...`、`GOWORK=off make boundary` 和 `GOWORK=off make schedulex-checks`。黄金文件覆盖触发器、DST、misfire；race gate 覆盖运行时并发。
