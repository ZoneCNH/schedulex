package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ZoneCNH/schedulex/pkg/schedulex"
)

func main() {
	s, _ := schedulex.NewScheduler()
	job := schedulex.JobFunc{NameValue: "heartbeat", RunFunc: func(context.Context) error {
		fmt.Println("tick")
		return nil
	}}
	_ = s.AddJob(job, schedulex.Every(time.Minute))
	fmt.Println(s.Snapshot().Version)
}
