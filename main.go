package main

import (
	"os"
	"os/signal"
	"context"
	"time"
	"strconv"
	
	qmq "github.com/rqure/qmq/src"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	ctx := context.Background()
	
	app := qmq.NewQMQApplication(ctx, "example")
	app.Initialize(ctx)
	defer app.Deinitialize(ctx)

	app.AddProducer(ctx, "clock:exchange", 10)

	tickRateMs, err := strconv.Atoi(os.Getenv("TICK_RATE_MS"))
	if err != nil {
		tickRateMs = 100
	}
	
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)	
	
	ticker := time.NewTicker(time.Duration(tickRateMs) * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			break
		case <-sigint:
			break
		case <-ticker.C:
			timestamp := qmq.QMQTimestamp{Value: timestamppb.Now()}
			app.Producer("clock:exchange").Push(ctx, &timestamp)
		}
	}
}
