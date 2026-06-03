# 测试策略母版

## 1. 定位

`xlib-standard` 是基础库测试策略母版，不绑定 `x.go` 业务模型。它定义所有生成基础库必须继承的测试分层、Gate 标准、Evidence 规则和扩展测试 profile。旧 `baselib-template` 名称只用于迁移说明。

## 2. 非目标

- 不引入完整 BDD 工具链。
- 不把 DDD 当作测试模式。
- 不默认强制 Chaos Test。
- 不默认强制 Mutation Test。
- 不默认运行长时间 Fuzz。
- 不绑定 `x.go` 业务模型。
- 不隐式读取生产密钥或 `/home/k8s/secrets/env/*`。

## 3. 测试分层

| Layer | 名称 | 内容 |
|---|---|---|
| L0 | Spec / ATDD | Spec、Acceptance Criteria、Traceability |
| L1 | Unit / TDD | Unit、Race、Lifecycle |
| L2 | Contract / Boundary / Security | Schema、API、Boundary、Secret、Vuln |
| L3 | Integration Smoke | Template render、Generated lib smoke |
| L4 | Property / Fuzz / Golden | Invariant、Fuzz smoke、Stable output |
| L5 | Compatibility / Observability | Error、Health、Metrics、JSON compatibility |
| L6 | Release Evidence | Manifest、Evidence、Review、Retrospective |
| L7 | Profile-Specific Heavy | Chaos、Mutation、Long Soak、Full E2E |

## 4. Required Gates

```text
GOWORK=off make fmt
GOWORK=off make vet
GOWORK=off make lint
GOWORK=off make test
GOWORK=off make race
GOWORK=off make boundary
GOWORK=off make security
GOWORK=off make contracts
GOWORK=off make docs-check
GOWORK=off make integration
GOWORK=off make evidence
GOWORK=off make release-check
```

## 5. Profile Gates

- Pure Library：`kernel`、`testkitx`。
- Config Library：`configx`。
- Observability Library：`observex`。
- Storage Library：`postgresx`、`redisx`、`taosx`、`ossx`、`clickhousex`。
- Messaging Library：`kafkax`。
