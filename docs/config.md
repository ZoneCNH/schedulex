# 配置模板

## 占位符

- `{{MODULE_NAME}}`
- `{{MODULE_PATH}}`
- `{{PACKAGE_NAME}}`

## 规则

- 配置必须由调用方显式传入。
- 不得隐式读取生产密钥目录。
- `Config` 必须支持 `Validate` 和 `Sanitize`。
- `Validate` 必须对空配置名和负数 timeout 返回 `ErrorKindValidation`。
- `contracts/config.schema.json` 的 `name`、`timeout_ms` 和 `secret` 必须与 `Config.Name`、`Config.Timeout` 和 `Config.Secret` 保持映射一致。
- 脱敏后的配置可以安全用于日志、Evidence 和发布说明。

生成的库可以在文档中说明由调用方拥有的配置层执行显式加载，然后只接收生成后的 `Config`。

本模板不得依赖 `x.go`。
