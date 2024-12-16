package main

import (
	"os"
	"time"

	"github.com/rqure/qlib/pkg/app"
	"github.com/rqure/qlib/pkg/app/workers"
	"github.com/rqure/qlib/pkg/data/store"
)

func getDatabaseAddress() string {
	addr := os.Getenv("Q_ADDR")
	if addr == "" {
		addr = "ws://webgateway:20000/ws"
	}

	return addr
}

func main() {
	s := store.NewWeb(store.WebConfig{
		Address: getDatabaseAddress(),
	})

	storeWorker := workers.NewStore(s)
	leadershipWorker := workers.NewLeadership(s)
	clockWorker := NewClockWorker(s, 1*time.Second)
	leadershipWorker.
		GetEntityFieldValidator().
		RegisterEntityFields("SystemClock", "CurrentTimeFn")

	storeWorker.Connected.Connect(leadershipWorker.OnStoreConnected)
	storeWorker.Disconnected.Connect(leadershipWorker.OnStoreDisconnected)

	leadershipWorker.BecameLeader().Connect(clockWorker.OnBecameLeader)
	leadershipWorker.LosingLeadership().Connect(clockWorker.OnLostLeadership)

	a := app.NewApplication("clock")
	a.AddWorker(storeWorker)
	a.AddWorker(leadershipWorker)
	a.AddWorker(clockWorker)
	a.Execute()
}
