package main

import (
	"context"
	"time"

	"github.com/rqure/qlib/pkg/app"
	"github.com/rqure/qlib/pkg/data"
	"github.com/rqure/qlib/pkg/data/query"
)

type ClockWorker struct {
	store    data.Store
	isLeader bool

	ticker *time.Ticker
}

func NewClockWorker(store data.Store, updateFrequency time.Duration) *ClockWorker {
	return &ClockWorker{
		store:    store,
		isLeader: false,
		ticker:   time.NewTicker(updateFrequency),
	}
}

func (w *ClockWorker) OnBecameLeader(context.Context) {
	w.isLeader = true
}

func (w *ClockWorker) OnLostLeadership(context.Context) {
	w.isLeader = false
}

func (w *ClockWorker) Init(context.Context, app.Handle) {
}

func (w *ClockWorker) Deinit(context.Context) {
	w.ticker.Stop()
}

func (w *ClockWorker) DoWork(ctx context.Context) {
	if !w.isLeader {
		return
	}

	select {
	case <-w.ticker.C:
		clocks := query.New(w.store).
			ForType("SystemClock").
			Execute(ctx)

		for _, clock := range clocks {
			clock.GetField("CurrentTimeFn").WriteTimestamp(ctx, time.Now())
		}
	default:

	}
}
