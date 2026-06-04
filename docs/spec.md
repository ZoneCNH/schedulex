# schedulex L1 规格

schedulex v0.1.0 提供确定性调度核心：触发器只根据输入时间计算下一次触发时间，调度器通过 `Clock` 获取时间，misfire、jitter、锁和并发策略均有合同测试覆盖。
