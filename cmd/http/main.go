package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/mirai-box/mirai-box/internal/app"
	"github.com/mirai-box/mirai-box/internal/config"
	"github.com/mirai-box/mirai-box/internal/logger"
)

func main() {
	conf, err := config.GetApplicationConfig()
	if err != nil {
		slog.Error("Can't get application config", "error", err)
		return
	}
	logger.Setup(conf)

	app, err := app.New(conf)
	if err != nil {
		slog.Error("Can't initialize application", "error", err)
		panic(err)
	}
	defer app.DB.Close()

	appliedMigrations, err := runMigrations(app.DB.DB, conf)
	if err != nil {
		slog.Error("Failed to migrate the database", "error", err)
		return
	} else {
		slog.Info("Migration completed", "appliedMigrations", appliedMigrations)
	}

	slog.Info("Start to listen on", "port", conf.Port)
	if err := http.ListenAndServe(":"+conf.Port, app.Router); err != nil {
		slog.Error("Failed to start server", "error", err)
	}
}

func runMigrations(instance *sql.DB, conf *config.Config) (int, error) {
	driver, err := postgres.WithInstance(instance, &postgres.Config{})
	if err != nil {
		return 0, fmt.Errorf("failed to create database driver: %w", err)
	}

	cwd := conf.ProjectRoot
	migrationPath := filepath.Join(cwd, "migrations")
	dbMigrationPath := fmt.Sprintf("file://%s", migrationPath)

	slog.Info("Running database migration",
		"migrationPath", migrationPath,
		"database", conf.Database.Database)

	m, err := migrate.NewWithDatabaseInstance(dbMigrationPath, conf.Database.Database, driver)
	if err != nil {
		return 0, fmt.Errorf("failed to create migration instance: %w", err)
	}

	// Get the current version before migration
	beforeVersion, _, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, fmt.Errorf("failed to get current migration version: %w", err)
	}

	// Run the migration
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return 0, fmt.Errorf("migration failed: %w", err)
	}

	// Get the version after migration
	afterVersion, _, err := m.Version()
	if err != nil {
		return 0, fmt.Errorf("failed to get new migration version: %w", err)
	}

	appliedMigrations := int(afterVersion - beforeVersion)

	if err == migrate.ErrNoChange {
		slog.Info("No migrations applied, database is up to date")
		return 0, nil
	}

	slog.Info("Database migration completed successfully",
		"appliedMigrations", appliedMigrations,
		"newVersion", afterVersion)

	return appliedMigrations, nil
}
