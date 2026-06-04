# 设计

核心边界分为触发器、时间源、调度运行时和外部锁接口。生产调度决策不直接调用 `time.Now` 或 `time.Sleep`；真实时间访问隔离在 `NewRealClock` 适配器中。
