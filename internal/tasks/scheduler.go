package tasks

import (
	"log"

	"github.com/hibiken/asynq"
)

// NewScheduler creates a periodic task manager that enqueues  tasks automatically on a cron schedule. It runs inside the worker process — not the HTTP server.
/*
┌─────────── minute      (0-59)
│ ┌───────── hour        (0-23)
│ │ ┌─────── day         (1-31)
│ │ │ ┌───── month       (1-12)
│ │ │ │ ┌─── weekday     (0-6, Sun=0 or SUN/MON/TUE...)
│ │ │ │ │
* * * * *

Examples:
"0 * * * *"      every hour at :00
 "*\/15 * * * *"   every 15 minutes
"0 9 * * MON"    every Monday at 9am
"0 0 * * *"      every day at midnight
"0 9 1 * *"      1st of every month at 9am
*/

func NewScheduler(redisAddr, redisPassword string, redisDB int) *asynq.Scheduler {
	scheduler := asynq.NewScheduler(
		asynq.RedisClientOpt{
			Addr:     redisAddr,
			Password: redisPassword,
			DB:       redisDB,
		},
		&asynq.SchedulerOpts{
			// Log scheduler errors
			EnqueueErrorHandler: func(task *asynq.Task, opts []asynq.Option, err error) {
				log.Printf("scheduler failed to enqueue task %q: %v", task.Type(), err)
			},
		},
	)

	// ── Register cron jobs ───────────────────────────────────────────────────

	// Cleanup expired OTPs — every hour
	if _, err := scheduler.Register(
		"0 * * * *",
		asynq.NewTask(TypeCleanupExpiredOTPs, nil,
			asynq.Queue("default"),
			asynq.MaxRetry(1),
		),
	); err != nil {
		log.Fatalf("failed to register cleanup otp task: %v", err)
	}

	// Weekly digest — every Monday at 9am
	if _, err := scheduler.Register(
		"0 9 * * MON",
		asynq.NewTask(TypeWeeklyDigest, nil,
			asynq.Queue("default"),
			asynq.MaxRetry(2),
		),
	); err != nil {
		log.Fatalf("failed to register weekly digest task: %v", err)
	}

	// Daily health report — every day at midnight
	if _, err := scheduler.Register(
		"0 0 * * *",
		asynq.NewTask(TypeDailyHealthReport, nil,
			asynq.Queue("default"),
			asynq.MaxRetry(1),
		),
	); err != nil {
		log.Fatalf("failed to register daily health report task: %v", err)
	}

	return scheduler
}
