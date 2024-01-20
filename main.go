package main

import (
	"context"
	qmq "github.com/rqure/qmq/src"
)

func main() {
	ctx := context.Background()
	app := qmq.NewQMQApplication(ctx, "example")
	app.Initialize(ctx)
	defer app.Deinitialize(ctx)
}
