package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ZoneCNH/schedulex/pkg/schedulex"
)

func main() {
	s := schedulex.NewScheduler(schedulex.Options{})
	_ = s.AddJob(schedulex.Job{ID: "heartbeat", Trigger: schedulex.Every(time.Minute), Run: func(context.Context) error { fmt.Println("tick"); return nil }})
	fmt.Println(s.Snapshot().Version)
}
