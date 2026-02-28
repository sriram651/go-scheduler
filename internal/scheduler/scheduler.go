package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	Schedule string
}

func New(schedule string) *Scheduler {
	return &Scheduler{
		Schedule: schedule,
	}
}

func (s *Scheduler) Start(ctx context.Context, job func()) {
	// Start the jobs here...
	c := cron.New(cron.WithLocation(time.Local))

	c.AddFunc(s.Schedule, job)

	log.Println("✅ Starting cron service...")

	c.Start()

	<-ctx.Done()

	<-c.Stop().Done()

	fmt.Println("Cron reminder service shutting down.")
}
