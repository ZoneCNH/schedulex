# schedulex 身份

## 我是谁

`schedulex` 是 **L1 运行时确定性任务调度库**，提供 trigger/clock/misfire/overlap/jitter 调度原语。

## 我做什么

| 能力 | 职责 |
|------|------|
| `trigger` | 调度触发策略（cron/interval/at） |
| `clock` | 可替换时钟接口（支持 FakeClock 测试） |
| `misfire` | 错失触发处理策略（skip/run_once/catch_up） |
| `overlap` | 任务重叠控制（allow/queue/skip） |
| `jitter` | 随机抖动，防惊群效应 |
| `EventSink` | 调度事件输出接口 |
| `Locker` | 分布式锁接口（实现由调用方注入） |

## 我不是什么

| 不是 | 原因 |
|------|------|
| **不是分布式锁实现** | Locker 是接口，实现由调用方提供 |
| **不是 exactly-once 语义引擎** | exactly-once 需要外部协调 |
| **不是业务任务编排引擎** | 业务语义由调用方定义 |
| **不是模板源** | 标准文档与生成脚本只描述基础库生成约束 |
| **不硬依赖 observex/resiliencx** | 观测和弹性策略由调用方组合注入 |

## 我的边界

```
我拥有:
  - trigger / clock / misfire / overlap / jitter 调度原语
  - EventSink 接口（事件输出契约）
  - Locker 接口（分布式锁契约）
  - snapshot（调度状态快照）
  - FakeClock（测试支持）

我不拥有:
  - 分布式锁实现（Redis/etcd/...）
  - 消息队列实现
  - 业务任务语义
  - 观测后端绑定
  - 弹性策略实现（属于 resiliencx）
```

## 宪法合规

| 条款 | 遵循方式 |
|------|----------|
| §1 P13 | 域内平级协作，调度原语之间无编译期依赖 |
| §2.1 | 明确声明拥有/不拥有 |
| §3 | 仅依赖 kernel（L0），观测/弹性通过接口注入 |
| §6 | 输出 EventSink 事件，交给调用方 |
