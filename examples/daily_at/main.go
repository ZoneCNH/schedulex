package main

import (
	"fmt"
	"time"

	"github.com/ZoneCNH/schedulex/pkg/schedulex"
)

func main() {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	next, _ := schedulex.DailyAt(9, 30, loc).Next(time.Date(2026, 6, 4, 8, 0, 0, 0, loc))
	fmt.Println(next.Format(time.RFC3339))
}
