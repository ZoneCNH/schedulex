package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ZoneCNH/schedulex/pkg/schedulex"
)

var heartbeatJob = schedulex.JobFunc{
	NameValue: "heartbeat",
	RunFunc: func(context.Context) error {
		fmt.Println("tick")
		return nil
	},
}

func main() {
	s, _ := schedulex.NewScheduler()
	_ = s.AddJob(heartbeatJob, schedulex.Every(time.Minute))
	fmt.Println(s.Snapshot().Version)
}
