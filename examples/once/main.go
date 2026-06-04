package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ZoneCNH/schedulex/pkg/schedulex"
)

func main() {
	now := time.Date(2026, 6, 4, 8, 0, 0, 0, time.UTC)
	s := schedulex.NewScheduler(schedulex.WithClock(schedulex.NewStaticClock(now)))
	_ = s.AddJob(schedulex.Job{ID: "once", Trigger: schedulex.Once(now.Add(time.Hour)), Run: func(context.Context) error { return nil }})
	fmt.Println(s.Snapshot().Jobs[0].Next.Format(time.RFC3339))
}
