package main

import (
	"os"

	"github.com/rqure/qlib/pkg/app"
	"github.com/rqure/qlib/pkg/app/workers"
	"github.com/rqure/qlib/pkg/data/store"
)

func getStoreAddress() string {
	addr := os.Getenv("Q_ADDR")
	if addr == "" {
		addr = "ws://localhost:20000/ws"
	}

	return addr
}

func main() {
	s := store.NewWeb(store.WebConfig{
		Address: getStoreAddress(),
	})

	storeWorker := workers.NewStore(s)
	leadershipWorker := workers.NewLeadership(s)
	scheduleWorker := NewScheduleWorker(s)

	validator := leadershipWorker.GetEntityFieldValidator()
	validator.RegisterEntityFields("SystemClock", "CurrentTimeFn")
	validator.RegisterEntityFields("Schedule", "CronExpression", "Loaded", "Enabled", "LastRun", "ExecuteFn")

	storeWorker.Connected.Connect(leadershipWorker.OnStoreConnected)
	storeWorker.Disconnected.Connect(leadershipWorker.OnStoreDisconnected)

	leadershipWorker.BecameLeader().Connect(scheduleWorker.OnBecameLeader)
	leadershipWorker.LosingLeadership().Connect(scheduleWorker.OnLostLeadership)

	a := app.NewApplication("clock")
	a.AddWorker(storeWorker)
	a.AddWorker(leadershipWorker)
	a.AddWorker(scheduleWorker)
	a.Execute()
}
