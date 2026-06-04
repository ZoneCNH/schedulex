package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ZoneCNH/schedulex/pkg/schedulex"
)

func main() {
	s := schedulex.NewScheduler(schedulex.Options{})
	_ = s.Start()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	fmt.Println(s.Shutdown(ctx) == nil && s.Shutdown(ctx) == nil)
}
