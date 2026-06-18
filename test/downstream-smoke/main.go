package main

import (
	"fmt"
	"time"

	"github.com/ZoneCNH/schedulex/pkg/schedulex"
)

func main() {
	next, _ := schedulex.Every(time.Minute).Next(time.Unix(0, 0).UTC())
	fmt.Println(next.Format(time.RFC3339))
}
