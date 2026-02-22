package main

import (
	"github.com/gin-gonic/gin"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/cache"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/config"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/handlers"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/middleware"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/utils"
	"gorm.io/gorm"

	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/repository"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/services"
)

// NewRouter wires together all dependencies and returns a configured *gin.Engine.
func NewRouter(db *gorm.DB, redis *cache.Client, cfg *config.Config) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())
	router.Use(middleware.RequestLogger())

	// --- Dependency wiring ---
	mailService := utils.NewMailService(cfg)
	userRepo := repository.NewUserRepository(db)
	otpRepo := repository.NewOTPRepository(redis)
	authSvc := services.NewAuthService(userRepo, otpRepo, mailService, cfg.JWTSecret, cfg.JWTExpiryHours)
	authHandler := handlers.NewAuthHandler(authSvc)
	healthHandler := handlers.NewHealthHandler(db, redis)

	// --- Shorthand so routes stay readable ---
	withTx := middleware.Transaction(db)
	withAuth := middleware.Auth(cfg.JWTSecret)

	// --- Health check ---
	router.GET("/health", healthHandler.Check)

	// --- API v1 ---
	v1 := router.Group("/api/v1")
	{
		// Public auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/signup", withTx, authHandler.Signup)
			auth.POST("/login", withTx, authHandler.Login)
			auth.POST("/send-verify-email-otp", authHandler.SendVerifyEmail)
			auth.POST("/verify-email", withTx, authHandler.VerifyEmail)
		}

		// Protected user routes (JWT required)
		users := v1.Group("/users")
		{
			users.GET("/", withAuth, authHandler.GetProfile)
			users.PUT("/", withAuth, withTx, authHandler.UpdateProfile)
			users.DELETE("/", withAuth, withTx, authHandler.DeleteAccount)

		}
	}

	return router
}
