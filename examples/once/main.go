package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ZoneCNH/schedulex/pkg/schedulex"
)

func main() {
	s, _ := schedulex.NewScheduler()
	job := schedulex.JobFunc{NameValue: "once", RunFunc: func(context.Context) error { return nil }}
	_ = s.AddJob(job, schedulex.Once(time.Now().Add(time.Hour)))
	fmt.Println(s.Snapshot().JobCount)
}
