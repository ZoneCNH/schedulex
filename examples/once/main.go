package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ZoneCNH/schedulex/pkg/schedulex"
)

var onceJob = schedulex.JobFunc{
	NameValue: "once",
	RunFunc:   func(context.Context) error { return nil },
}

func main() {
	s, _ := schedulex.NewScheduler()
	_ = s.AddJob(onceJob, schedulex.Once(time.Now().Add(time.Hour)))
	fmt.Println(s.Snapshot().JobCount)
}
