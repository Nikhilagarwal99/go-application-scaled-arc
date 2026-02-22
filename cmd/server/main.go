package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/cache"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/config"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/database"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/logger"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/server"
	"go.uber.org/zap"
)

func main() {
	// 1. Load config
	cfg := config.Load()

	// 2. Logger MUST be first — everything after can use it safely
	logger.Init(cfg.AppEnv)
	defer logger.Sync()

	// 3. Connect to Redis
	redis := cache.New(cfg)

	// 4. Connect to PostgreSQL + run migrations
	db := database.Connect(cfg)
	database.Migrate(db)

	// 5. Boot the router with all dependencies
	router := server.NewRouter(db, redis, cfg)

	// 6. Configure HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 7. Start in background goroutine
	go func() {
		logger.Info("server started", zap.String("port", cfg.ServerPort), zap.String("env", cfg.AppEnv))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed to start", zap.Error(err))
		}
	}()

	// 8. Block until SIGINT or SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server gracefully")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("forced shutdown", zap.Error(err))
	}

	logger.Info("server exited cleanly")
}
