package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ZoneCNH/schedulex/pkg/schedulex"
)

func main() {
	s := schedulex.NewScheduler(schedulex.Options{})
	_ = s.AddJob(schedulex.Job{ID: "once", Trigger: schedulex.Once(time.Now().Add(time.Hour)), Run: func(context.Context) error { return nil }})
	fmt.Println(s.Snapshot().JobCount)
}
