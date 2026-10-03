package main

import (
	"context"
	"fmt"
	"os"

	"github.com/stuffstash/stuff-stash/cmd/label-catalog/internal/bootstrap"
)

func main() {
	if err := bootstrap.Run(context.Background(), os.Args[1:], os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
