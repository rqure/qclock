package main

import (
	"context"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/rqure/qlib/pkg/app"
	"github.com/rqure/qlib/pkg/data"
	"github.com/rqure/qlib/pkg/data/notification"
	"github.com/rqure/qlib/pkg/data/query"
	"github.com/rqure/qlib/pkg/log"
)

type ScheduleWorker struct {
	store     data.Store
	isLeader  bool
	scheduler *gocron.Scheduler
	tokens    []data.NotificationToken
}

func NewScheduleWorker(store data.Store) *ScheduleWorker {
	return &ScheduleWorker{
		store:     store,
		isLeader:  false,
		scheduler: gocron.NewScheduler(time.UTC),
		tokens:    []data.NotificationToken{},
	}
}

func (w *ScheduleWorker) OnBecameLeader(ctx context.Context) {
	w.isLeader = true

	// Setup notifications for schedule changes
	for _, token := range w.tokens {
		token.Unbind(ctx)
	}
	w.tokens = []data.NotificationToken{}

	w.tokens = append(w.tokens,
		w.store.Notify(ctx,
			notification.NewConfig().
				SetEntityType("Schedule").
				SetFieldName("CronExpression"),
			notification.NewCallback(w.onScheduleChanged)),
		w.store.Notify(ctx,
			notification.NewConfig().
				SetEntityType("Schedule").
				SetFieldName("Enabled"),
			notification.NewCallback(w.onScheduleChanged)),
		w.store.Notify(ctx,
			notification.NewConfig().
				SetEntityType("Root").
				SetFieldName("SchemaUpdateTrigger"),
			notification.NewCallback(w.onSchemaUpdated)),
	)

	w.loadSchedules(ctx)
}

func (w *ScheduleWorker) OnLostLeadership(ctx context.Context) {
	w.isLeader = false
	w.scheduler.Stop()

	for _, token := range w.tokens {
		token.Unbind(ctx)
	}
	w.tokens = []data.NotificationToken{}
}

func (w *ScheduleWorker) Init(context.Context, app.Handle) {
}

func (w *ScheduleWorker) Deinit(context.Context) {
	w.scheduler.Stop()
}

func (w *ScheduleWorker) DoWork(ctx context.Context) {
	if !w.isLeader {
		return
	}
}

func (w *ScheduleWorker) onScheduleChanged(ctx context.Context, n data.Notification) {
	log.Info("Schedule changed: %s.%s", n.GetCurrent().GetEntityId(), n.GetCurrent().GetFieldName())

	if w.isLeader {
		w.loadSchedules(ctx)
	}
}

func (w *ScheduleWorker) onSchemaUpdated(ctx context.Context, n data.Notification) {
	if w.isLeader {
		w.loadSchedules(ctx)
	}
}

func (w *ScheduleWorker) loadSchedules(ctx context.Context) {
	log.Info("Loading all enabled schedules...")

	w.scheduler.Stop()
	w.scheduler.Clear()

	schedules := query.New(w.store).
		Select("CronExpression").
		From("Schedule").
		Where("Enabled").Equals(true).
		Execute(ctx)

	for _, schedule := range schedules {
		cronExpr := schedule.GetField("CronExpression").GetString()

		_, err := w.scheduler.CronWithSeconds(cronExpr).Do(func() {
			log.Info("Executing schedule: %s (%s)", schedule.GetName(), schedule.GetId())
			schedule.DoMulti(ctx, func(schedule data.EntityBinding) {
				schedule.GetField("ExecuteFn").WriteInt(ctx)
				schedule.GetField("LastRun").WriteTimestamp(ctx, time.Now())
			})
		})

		if err != nil {
			schedule.GetField("Loaded").WriteBool(ctx, false, data.WriteChanges)
			log.Error("Failed to load schedule: %s (%s): %s", schedule.GetName(), schedule.GetId(), err)
		} else {
			schedule.GetField("Loaded").WriteBool(ctx, true, data.WriteChanges)
			log.Info("Loaded schedule: %s (%s)", schedule.GetName(), schedule.GetId())
		}
	}

	w.scheduler.StartAsync()
}
