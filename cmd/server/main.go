package main

import (
	"context"
	"log"
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

	// 1. Load configuration
	cfg := config.Load()

	// initialise logger before anything else
	logger.Init(cfg.AppEnv)
	defer logger.Sync() // flush on shutdown

	// 2. Connect to database and run migrations
	redis := cache.New(cfg)
	db := database.Connect(cfg)
	database.Migrate(db)

	// 3. Build the HTTP router
	router := server.NewRouter(db, redis, cfg)

	// 4. Create the HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 5. Start serving in a goroutine so we can listen for shutdown signals
	go func() {
		logger.Info("starting server", zap.String("port", cfg.ServerPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("failed to start server", zap.Error(err))
		}
	}()

	// 6. Graceful shutdown — wait for SIGINT / SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("Server exited gracefully")
}
