package main

import (
	"context"
	qmq "qmq/src"
)

func main() {
  ctx := context.Background()
	app := qmq.NewQMQApplication(ctx, "example")
	app.Initialize(ctx)
	defer app.Deinitialize(ctx)
}
