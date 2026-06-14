# schedulex 发布治理宪章

本仓库是 `github.com/ZoneCNH/schedulex` L1 deterministic scheduler 的可审计源。所有发布、治理和证据相关命令默认使用 `GOWORK=off`，并以可复现 contract、fixture、manifest 和 checksum 作为发布依据。

1. `docs/standard/`、`docs/spec.md`、`docs/api.md` 和 `docs/release.md` 描述人类可读的调度器契约与发布流程。
2. `contracts/` 保存公共 API snapshot、golden cases、schema 和 release manifest contract；版本元数据必须与 `pkg/schedulex.Version`、README 和 release docs 一致。
3. `scripts/` 与 `Makefile` 描述机器可执行的 release gates，包括 docs、boundary、contract、downstream smoke、manifest 和 preflight 校验。
4. `release/manifest/` 与 `release/downstream-adoption/` 保存发布证据模板、fixture 或生成产物；`latest.json` 与 `.sha256` 必须成对更新并可校验。

任何 release 变更必须先更新 contract 与文档，再运行相应门禁。不得用占位文件伪造下游采用或远端 module fetch 通过证据；发布前可以声明 fixture no-local-replace contract 通过，发布后必须用远端网络模式复核 tag 可解析性。
