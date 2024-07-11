package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/mirai-box/mirai-box/internal/app"
	"github.com/mirai-box/mirai-box/internal/config"
	"github.com/mirai-box/mirai-box/internal/database"
	"github.com/mirai-box/mirai-box/internal/logger"
)

func main() {
	// Load configuration
	slog.Info("Load service configuration")
	conf, err := config.GetApplicationConfig()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Setup logger
	logger.Setup(conf)

	// Ensure database and user are created
	if err := database.CreateDatabaseAndUser(conf); err != nil {
		slog.Error("Failed to create database and user", "error", err)
		os.Exit(1)
	}

	// Database connection
	dbConnectionString := conf.Database.ConnectionString()
	slog.Info("database connection string", "dsn", dbConnectionString)
	db, err := gorm.Open(postgres.Open(dbConnectionString), &gorm.Config{})
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Initialize the app
	router := app.SetupRoutes(db, conf)

	// Start the server
	slog.Info("Starting server", "port", conf.Port)
	err = http.ListenAndServe(":"+conf.Port, router)
	if err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
