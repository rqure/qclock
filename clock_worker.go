package main

import (
	"time"

	qdb "github.com/rqure/qdb/src"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ClockWorker struct {
	db       qdb.IDatabase
	isLeader bool

	ticker *time.Ticker
}

func NewClockWorker(db qdb.IDatabase, updateFrequency time.Duration) *ClockWorker {
	return &ClockWorker{
		db:       db,
		isLeader: false,
		ticker:   time.NewTicker(updateFrequency),
	}
}

func (w *ClockWorker) OnBecameLeader() {
	w.isLeader = true
}

func (w *ClockWorker) OnLostLeadership() {
	w.isLeader = false
}

func (w *ClockWorker) Init() {
}

func (w *ClockWorker) Deinit() {
	w.ticker.Stop()
}

func (w *ClockWorker) DoWork() {
	if !w.isLeader {
		return
	}

	select {
	case <-w.ticker.C:
		clocks := qdb.NewEntityFinder(w.db).Find(qdb.SearchCriteria{
			EntityType: "SystemClock",
		})

		for _, clock := range clocks {
			clock.GetField("CurrentTimeFn").PushValue(&qdb.Timestamp{Raw: timestamppb.Now()})
		}
	default:

	}
}
