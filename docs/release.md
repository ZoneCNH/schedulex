# 发布

v0.1.0 发布前运行 `GOWORK=off make release-preflight VERSION=v0.1.0`，并至少保留以下证据：

- `GOWORK=off make release-final-check`
- `release/manifest/latest.json`
- `release/manifest/latest.json.sha256`

下游 smoke 当前只验证 fixture 不使用本地 `replace`；完整远端 module fetch 需要 `github.com/ZoneCNH/schedulex v0.1.0` 已发布。
