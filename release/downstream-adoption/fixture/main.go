package main

import (
	"context"
	"time"

	"github.com/ZoneCNH/schedulex/pkg/schedulex"
)

func main() {
	s, _ := schedulex.NewScheduler(schedulex.WithClock(schedulex.NewStaticClock(time.Unix(0, 0))))
	job := schedulex.JobFunc{NameValue: "smoke", RunFunc: func(context.Context) error { return nil }}
	_ = s.AddJob(job, schedulex.Every(time.Minute))
	_ = s.Snapshot()
}
