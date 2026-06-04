package main

import (
	"fmt"
	"time"

	"github.com/ZoneCNH/schedulex/pkg/schedulex"
)

func main() {
	tr, _ := schedulex.Cron("*/20 * * * *")
	next, _ := tr.Next(time.Date(2026, 6, 4, 8, 7, 0, 0, time.UTC))
	fmt.Println(next.Format(time.RFC3339))
}
