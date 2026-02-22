package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

func Connect(cfg *config.Config) *gorm.DB {
	masterDSN := buildDSN(
		cfg.DBHost, cfg.DBPort, cfg.DBUser,
		cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	slaveDSN := buildDSN(
		cfg.DBSlaveHost, cfg.DBSlavePort, cfg.DBSlaveUser,
		cfg.DBSlavePassword, cfg.DBSlaveName, cfg.DBSlaveSSLMode,
	)

	gormLogger := logger.Default.LogMode(logger.Silent)
	if cfg.AppEnv == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(masterDSN), &gorm.Config{
		Logger:            gormLogger,
		AllowGlobalUpdate: false,
	})
	if err != nil {
		log.Fatalf("failed to connect to master database: %v", err)
	}

	err = db.Use(dbresolver.Register(dbresolver.Config{
		Sources:           []gorm.Dialector{postgres.Open(masterDSN)},
		Replicas:          []gorm.Dialector{postgres.Open(slaveDSN)},
		Policy:            dbresolver.RandomPolicy{},
		TraceResolverMode: true,
	}))
	if err != nil {
		log.Fatalf("failed to register dbresolver: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get generic database object: %v", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(1 * time.Minute)

	log.Println("database connected")
	return db
}

func Migrate(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("migration: failed to get sql.DB: %v", err)
	}

	driver, err := migratepg.WithInstance(sqlDB, &migratepg.Config{})
	if err != nil {
		log.Fatalf("migration: failed to create driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/database/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatalf("migration: failed to init: %v", err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("migrations: already up to date")
			return
		}
		log.Fatalf("migration: failed to run: %v", err)
	}

	log.Println("migrations: all pending migrations applied")
}

func MigrateDown(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("migration: failed to get sql.DB: %v", err)
	}

	driver, err := migratepg.WithInstance(sqlDB, &migratepg.Config{})
	if err != nil {
		log.Fatalf("migration: failed to create driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/database/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatalf("migration: failed to initialise: %v", err)
	}

	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("migration: failed to roll back: %v", err)
	}

	log.Println("migrations: rolled back successfully")
}

func Ping(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("postgres unreachable: %w", err)
	}

	return nil
}

func buildDSN(host, port, user, password, dbname, sslmode string) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		host, port, user, password, dbname, sslmode,
	)
}
