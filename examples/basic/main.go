package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/ZoneCNH/xlib-standard/pkg/templatex"
)

func main() {
	run(os.Stdout, os.Stderr, templatex.Config{Name: "templatex"})
}

func run(stdout, stderr io.Writer, cfg templatex.Config) {
	client, err := templatex.New(context.Background(), cfg)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "create client: %v\n", err)
		return
	}
	defer func() {
		_ = client.Close(context.Background())
	}()

	_, _ = fmt.Fprintln(stdout, templatex.ModuleName)
}
