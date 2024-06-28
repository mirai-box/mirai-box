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

	if err := runMigrations(app.DB.DB, conf); err != nil {
		slog.Error("Failed to migrate the database", "error", err)
		return
	}

	slog.Info("Start to listen on", "port", conf.Port)
	if err := http.ListenAndServe(":"+conf.Port, app.Router); err != nil {
		slog.Error("Failed to start server", "error", err)
	}
}

func runMigrations(instance *sql.DB, conf *config.Config) error {
	driver, err := postgres.WithInstance(instance, &postgres.Config{})
	if err != nil {
		return err
	}

	cwd := conf.ProjectRoot

	migrationPath := filepath.Join(cwd, "migrations")

	slog.Info("run database migration", "migrationPath", migrationPath, "database", conf.Database.Database)

	dbMigrationPath := fmt.Sprintf("file://%s", migrationPath)

	m, err := migrate.NewWithDatabaseInstance(dbMigrationPath, conf.Database.Database, driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
