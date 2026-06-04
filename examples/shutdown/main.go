package main

import (
	"context"
	"fmt"

	"github.com/ZoneCNH/schedulex/pkg/schedulex"
)

func main() {
	s := schedulex.NewScheduler()
	_ = s.Shutdown(context.Background())
	_ = s.Shutdown(context.Background())
	fmt.Println(s.Snapshot().Shutdown)
}
