# 发布模板

> 模板说明：本文件是发布证据填写模板，不是 DONE 证据本身。只有填写了
> workflow run、artifact URL、sha256、score 和 gate 结果的新鲜实例才可作为发布证据。

## 版本

- 版本:
- Commit:
- Tree SHA:

## 必需 Evidence

- [ ] `GOWORK=off make release-check`
- [ ] `go run ./cmd/goalcli score --min 9.8`
- [ ] `GOWORK=off make release-final-check`
- [ ] `GOWORK=off make release-preflight VERSION=<version>`
- [ ] `release/manifest/latest.json` 已生成并完成校验
- [ ] `release/manifest/latest.json.sha256` 已生成并完成校验

## 分数

- Score JSON:
- 阈值:
- 状态:

## 备注

- 模板行为影响:
- 生成库影响:
- 已知缺口:

## Artifact

- CI artifact:
- Workflow run ID:
- Artifact URL:
- 本地 manifest:
