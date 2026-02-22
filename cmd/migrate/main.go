package main

import (
	"errors"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/config"
	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/database"
)

/*
# install the CLI once
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# create a new migration
migrate create -ext sql -dir migrations -seq add_phone_to_users


This generates:

migrations/000002_add_phone_to_users.up.sql
migrations/000002_add_phone_to_users.down.sql

migrate create -ext sql -dir internal/database/migrations -seq your_migration_name
# fills in the version number automatically
# then write your SQL in the generated files
*/

func main() {
	// Expect exactly one argument — "up" or "down"
	if len(os.Args) < 2 {
		log.Fatal("usage: migrate <up|down>")
	}

	command := os.Args[1]
	if command != "up" && command != "down" {
		log.Fatalf("unknown command %q — use 'up' or 'down'", command)
	}

	// Load config and connect to DB
	cfg := config.Load()
	db := database.Connect(cfg)

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB: %v", err)
	}

	driver, err := migratepg.WithInstance(sqlDB, &migratepg.Config{})
	if err != nil {
		log.Fatalf("failed to create migrate driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/database/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatalf("failed to initialise migrate: %v", err)
	}

	// Show current version before running
	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		log.Fatalf("failed to get migration version: %v", err)
	}
	if errors.Is(err, migrate.ErrNilVersion) {
		log.Println("current migration version: none (fresh database)")
	} else {
		log.Printf("current migration version: %d (dirty: %v)", version, dirty)
	}

	// Run the command
	switch command {
	case "up":
		runUp(m)
	case "down":
		runDown(m)
	}
}

func runUp(m *migrate.Migrate) {
	log.Println("running migrations up...")

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("nothing to migrate — already up to date")
			return
		}
		log.Fatalf("migration up failed: %v", err)
	}

	version, _, _ := m.Version()
	log.Printf("migrations complete — now at version %d", version)
}

func runDown(m *migrate.Migrate) {
	// Safety guard — down in production should be a very deliberate act
	if os.Getenv("APP_ENV") == "production" {
		log.Fatal("migrate down is disabled in production — set APP_ENV != production to proceed")
	}

	log.Println("rolling back all migrations...")

	if err := m.Down(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("nothing to roll back — already at zero")
			return
		}
		log.Fatalf("migration down failed: %v", err)
	}

	log.Println("all migrations rolled back")
}
