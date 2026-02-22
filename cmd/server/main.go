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
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/server"
)

func main() {
	// 1. Load configuration
	cfg := config.Load()

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
		log.Printf("🚀 Server running on port %s [%s]", cfg.ServerPort, cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
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
