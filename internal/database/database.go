package database

import (
	"fmt"
	"log"

	"github.com/nikhilAgarwal99/goapp/internal/config"
	"github.com/nikhilAgarwal99/goapp/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect opens a connection pool to PostgreSQL and returns a *gorm.DB.
func Connect(cfg *config.Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	gormLogger := logger.Default.LogMode(logger.Silent)
	if cfg.AppEnv == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get generic database object: %v", err)
	}

	// Connection pool settings — tune for your workload
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	log.Println("Database connection established")
	return db
}

// Migrate runs auto-migration for all registered models.
func Migrate(db *gorm.DB) {
	if err := db.AutoMigrate(
		&models.User{},
	); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}
	log.Println("Database migration completed")
}
