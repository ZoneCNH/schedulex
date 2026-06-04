package main

import (
	"fmt"
	"time"

	"github.com/ZoneCNH/schedulex/pkg/schedulex"
)

func main() {
	base := time.Date(2026,1,1,0,0,0,0,time.UTC)
	d := schedulex.PlanMisfire(schedulex.MisfireRunOnce, []time.Time{base, base.Add(time.Minute)}, time.Time{}, false)
	fmt.Println(len(d.Runs))
}
