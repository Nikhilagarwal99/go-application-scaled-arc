package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres" // aliased
	_ "github.com/golang-migrate/migrate/v4/source/file"               // aliased
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

// Connect opens a connection pool to PostgreSQL and returns a *gorm.DB.
func Connect(cfg *config.Config) *gorm.DB {
	masterDSN := buildDSN(
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBSSLMode,
	)

	slaveDSN := buildDSN(
		cfg.DBSlaveHost,
		cfg.DBSlavePort,
		cfg.DBSlaveUser,
		cfg.DBSlavePassword,
		cfg.DBSlaveName,
		cfg.DBSlaveSSLMode,
	)

	gormLogger := logger.Default.LogMode(logger.Silent)
	if cfg.AppEnv == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	// Open connection using master DSN
	db, err := gorm.Open(postgres.Open(masterDSN), &gorm.Config{
		Logger: gormLogger,

		// Prevents any accidental updates without a WHERE clause
		// e.g db.Delete(&User{}) without condition will return error
		AllowGlobalUpdate: false,
	})
	if err != nil {
		log.Fatalf("failed to connect to master database: %v", err)
	}

	// Register dbresolver — this is where read/write splitting happens.
	// GORM intercepts every query and routes it based on operation type.
	err = db.Use(dbresolver.Register(dbresolver.Config{
		// Sources = masters — all writes go here
		Sources: []gorm.Dialector{
			postgres.Open(masterDSN),
		},

		// Replicas = slaves — all reads go here
		// Add more postgres.Open(anotherSlaveDSN) to distribute further
		Replicas: []gorm.Dialector{
			postgres.Open(slaveDSN),
		},

		// Round robin across replicas if you have multiple slaves
		Policy: dbresolver.RandomPolicy{},

		// When true — reads inside a transaction always go to master.
		// This is critical — slave replication has a small lag,
		// so reading from slave mid-transaction could return stale data.
		TraceResolverMode: true,
	}))
	if err != nil {
		log.Fatalf("failed to register dbresolver: %v", err)
	}

	// Configure connection pool on the underlying sql.DB
	// dbresolver manages pools for each source/replica separately
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

// Migrate runs all pending migrations from the migrations/ directory.
// It uses golang-migrate which tracks versions in the schema_migrations table.
// This replaces AutoMigrate entirely.
func Migrate(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("migration: failed to get sql.DB: %v", err)
	}

	// Register the postgres driver with golang-migrate
	driver, err := migratepg.WithInstance(sqlDB, &migratepg.Config{})
	if err != nil {
		log.Fatalf("migration: failed to create driver: %v", err)
	}

	// Point golang-migrate at our migrations directory
	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/database/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatalf("migration: failed to init: %v", err)
	}

	// Run all pending up migrations
	if err := m.Up(); err != nil {
		// ErrNoChange is not a real error — it just means
		// all migrations have already been applied
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("migrations: already up to date")
			return
		}
		log.Fatalf("migration: failed to run: %v", err)
	}

	log.Println("migrations: all pending migrations applied")
}

// MigrateDown rolls back all migrations — useful in tests only.
// Never call this in production code.
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

	log.Println("Database migrations rolled back")
}

// Ping checks whether the master DB connection is alive.
// Called by the health handler on every /health request.
// Uses a short timeout so a hung DB doesn't block the health check forever.
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
