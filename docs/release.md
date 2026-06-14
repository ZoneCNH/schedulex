# 发布

v1.0.0 发布前运行 `GOWORK=off make release-preflight VERSION=v1.0.0`，并至少保留以下证据：

- `release/downstream-adoption/latest.json`
- `release/downstream-adoption/latest.json.sha256`
- `release/manifest/latest.json`
- `release/manifest/latest.json.sha256`

下游 smoke 当前只验证 fixture 不使用本地 `replace`；完整远端 module fetch 需要 `github.com/ZoneCNH/schedulex v1.0.0` 已发布。发布前的本地声明必须明确区分：fixture contract 已通过，远端 module fetch 需在 tag 发布后用 `SCHEDULEX_DOWNSTREAM_NETWORK=1 ./scripts/check_downstream_smoke.sh` 复核。

最终完成声明必须包含 `DONE with evidence:`，并列出 preflight、contract tests、downstream smoke、manifest checksum 的实际命令输出。
