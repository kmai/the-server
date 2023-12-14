package main

import (
	"context"

	"github.com/kmai/the-server/internal/router"
)

func main() {
	ctx := context.Background()
	router.Start(ctx)
}
