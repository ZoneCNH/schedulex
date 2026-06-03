# Security 规则

> 源自 Goal 完整规则 v1.0 §21

## RULE-SECURITY-001：禁止提交密钥

禁止进入：

```text
source code
README
tests
logs
release manifest
PR description
evidence report
```

## RULE-SECURITY-002：密钥必须走外部注入

对于基础库体系，默认使用：

```text
/home/k8s/secrets/env/*
```

## RULE-SECURITY-003：Evidence 不能泄漏敏感信息

Evidence 中必须过滤：

```text
token
password
secret
private key
access key
cookie
authorization header
```
