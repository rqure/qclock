package main

import (
	"os"
	"os/signal"
	"strconv"
	"time"

	qmq "github.com/rqure/qmq/src"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	app := qmq.NewQMQApplication("clock")
	app.Initialize()
	defer app.Deinitialize()

	app.AddProducer("clock:exchange").Initialize(10)

	tickRateMs, err := strconv.Atoi(os.Getenv("TICK_RATE_MS"))
	if err != nil {
		tickRateMs = 100
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	ticker := time.NewTicker(time.Duration(tickRateMs) * time.Millisecond)
	for {
		select {
		case <-sigint:
			app.Logger().Advise("SIGINT received")
			return
		case <-ticker.C:
			timestamp := qmq.QMQTimestamp{Value: timestamppb.Now()}
			app.Logger().Advise("Tick")
			app.Producer("clock:exchange").Push(&timestamp)
		}
	}
}
