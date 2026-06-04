package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ZoneCNH/schedulex/pkg/schedulex"
)

func main() {
	s, _ := schedulex.NewScheduler()
	_ = s.Start(context.Background())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	first := s.Shutdown(ctx)
	second := s.Shutdown(ctx)
	fmt.Println(first == nil && second == nil)
}
