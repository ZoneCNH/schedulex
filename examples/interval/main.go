package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ZoneCNH/schedulex/pkg/schedulex"
)

func main() {
	clock := schedulex.NewStaticClock(time.Date(2026, 6, 4, 8, 0, 0, 0, time.UTC))
	s := schedulex.NewScheduler(schedulex.WithClock(clock))
	_ = s.AddJob(schedulex.Job{ID: "interval", Trigger: schedulex.Every(15 * time.Minute), Run: func(context.Context) error { return nil }})
	fmt.Println(s.Snapshot().Jobs[0].Next.Format(time.RFC3339))
}
