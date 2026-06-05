# Evidence

Full Goal Runtime v3.1 的完成 Evidence 必须区分 required、extended、final 和下游集成 gate。

## Required Evidence

- `GOWORK=off go test ./...`：通过。
- `GOWORK=off make boundary`：通过。
- `GOWORK=off make contracts`：通过。
- `GOWORK=off make docs-check`：通过。
- `GOWORK=off make security`：`govulncheck` 和 secret scan 均通过。
- `GOWORK=off make release-check`：通过，并生成 `release/manifest/latest.json` 与 `.sha256`。

## Final Evidence

- `GOWORK=off make release-final-check`：通过且工作区 clean。
- `GOWORK=off make release-preflight VERSION=<version>`：通过。
- `goalcli score --min 9.8`：通过。
- kernel downstream smoke：通过。

## Declaration

最终声明必须使用：

```text
DONE with evidence:
```

Evidence、manifest、PR 描述、日志和复盘不得包含 `/home/k8s/secrets/env/*` 下任何文件内容。
