package main

import (
	"fmt"
	"time"

	"github.com/ZoneCNH/schedulex/pkg/schedulex"
)

func main() {
	missed := []time.Time{time.Unix(1, 0), time.Unix(2, 0)}
	fmt.Println(len(schedulex.ReconcileMisfire(schedulex.MisfireRunOnce, missed)))
}
