package server

import (
	"github.com/gin-gonic/gin"
	"github.com/nikhilAgarwal99/goapp/internal/cache"
	"github.com/nikhilAgarwal99/goapp/internal/config"
	"github.com/nikhilAgarwal99/goapp/internal/handlers"
	"github.com/nikhilAgarwal99/goapp/internal/middleware"
	"github.com/nikhilAgarwal99/goapp/internal/utils"
	"gorm.io/gorm"

	"github.com/nikhilAgarwal99/goapp/internal/repository"
	"github.com/nikhilAgarwal99/goapp/internal/services"
)

// NewRouter wires together all dependencies and returns a configured *gin.Engine.
func NewRouter(db *gorm.DB, redis *cache.Client, cfg *config.Config) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(gin.Logger()) // replace with a structured logger (e.g. zap) for production

	// --- Dependency wiring ---
	mailService := utils.NewMailService(cfg)
	userRepo := repository.NewUserRepository(db)
	otpRepo := repository.NewOTPRepository(redis)
	authSvc := services.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiryHours, otpRepo, mailService)
	authHandler := handlers.NewAuthHandler(authSvc)

	// --- Health check ---
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// --- API v1 ---
	v1 := router.Group("/api/v1")
	{
		// Public auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/signup", authHandler.Signup)
			auth.POST("/login", authHandler.Login)
			auth.POST("/send-verify-email-otp", authHandler.SendVerifyEmail)
			auth.POST("/verify-email", authHandler.VerifyEmail)
		}

		// Protected user routes (JWT required)
		users := v1.Group("/users")
		users.Use(middleware.Auth(cfg.JWTSecret))
		{
			users.GET("/", authHandler.GetProfile)
			users.PUT("/", authHandler.UpdateProfile)
			users.DELETE("/", authHandler.DeleteAccount)

		}
	}

	return router
}
