package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/config"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/logger"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/tasks"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/utils"
	"go.uber.org/zap"
)

func main() {
	// 1. Load config
	cfg := config.Load()

	// 2. Init logger before everything else
	logger.Init(cfg.AppEnv)
	defer logger.Sync()

	// 3. Build mail service — worker needs it to send emails
	mailService := utils.NewMailService(cfg)

	// 4. Build the email processor — holds all task handler functions
	emailProcessor := tasks.NewEmailProcessor(mailService)

	// 5. Create asynq server
	// This is where you configure concurrency and queue priorities
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		},
		asynq.Config{
			// How many tasks to process concurrently
			// Rule of thumb: number of CPUs * 2 for IO-bound tasks like email
			Concurrency: 10,

			// Queue priorities — critical is processed 3x more than default
			// Numbers are relative weights not percentages
			Queues: map[string]int{
				"critical": 3, // verify email — user is waiting
				"default":  1, // welcome email — background
			},

			// Log asynq internal errors using our zap logger
			ErrorHandler: asynq.ErrorHandlerFunc(
				func(ctx context.Context, task *asynq.Task, err error) {
					logger.Error("task failed after all retries",
						zap.String("type", task.Type()),
						zap.Error(err),
					)
				},
			),
		},
	)

	// 6. Register task handlers
	// mux maps task type strings to their processor functions
	// Same concept as HTTP router — task type = route, processor = handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeVerifyEmail, emailProcessor.ProcessVerifyEmail)
	mux.HandleFunc(tasks.TypeWelcomeEmail, emailProcessor.ProcessWelcomeEmail)

	// Extra tasks Scheduled Cron Jobs
	mux.HandleFunc(tasks.TypeCleanupExpiredOTPs, emailProcessor.ProcessCleanupExpiredOTPs)
	mux.HandleFunc(tasks.TypeWeeklyDigest, emailProcessor.ProcessWeeklyDigest)
	mux.HandleFunc(tasks.TypeDailyHealthReport, emailProcessor.ProcessDailyHealthReport)

	// ── Scheduler — enqueues cron jobs automatically ─────────────────────────
	scheduler := tasks.NewScheduler(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)

	// 7. -----Start both worker and scheduler ─────────────────────────────────────
	go func() {
		logger.Info("worker started",
			zap.String("queues", "critical, default"),
			zap.Int("concurrency", 10),
		)
		if err := srv.Run(mux); err != nil {
			log.Fatalf("worker failed to start: %v", err)
		}
	}()

	go func() {
		logger.Info("scheduler started")
		if err := scheduler.Run(); err != nil {
			log.Fatalf("scheduler failed to start: %v", err)
		}
	}()

	// 8. Block until SIGINT or SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 9. Graceful shutdown
	// Waits for in-progress tasks to finish before stopping
	logger.Info("shutting down worker and scheduler gracefully")
	scheduler.Shutdown()
	srv.Shutdown()
	logger.Info("worker exited cleanly")
}
