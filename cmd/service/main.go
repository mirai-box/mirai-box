package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/mirai-box/mirai-box/internal/app"
	"github.com/mirai-box/mirai-box/internal/config"
	"github.com/mirai-box/mirai-box/internal/logger"
	"github.com/mirai-box/mirai-box/internal/models"
)

func createDatabaseAndUser(conf *config.Config) error {
	conn, err := sqlx.Connect("postgres", conf.Database.ConnectionPostgres())
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	defer conn.Close()

	// Create user
	query := fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s'",
		conf.Database.Username, conf.Database.Password)
	_, err = conn.Exec(query)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Create database
	query = fmt.Sprintf("CREATE DATABASE %s OWNER %s", conf.Database.Database, conf.Database.Username)
	_, err = conn.Exec(query)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to create database: %w", err)
	}

	// Grant privileges
	query = fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE %s TO %s", conf.Database.Database, conf.Database.Username)
	_, err = conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to grant privileges: %w", err)
	}

	query = fmt.Sprintf("GRANT ALL PRIVILEGES ON SCHEMA public TO %s", conf.Database.Username)
	if _, err = conn.Exec(query); err != nil {
		return fmt.Errorf("failed to grant privileges: %w", err)
	}

	query = fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO %s", conf.Database.Username)
	if _, err = conn.Exec(query); err != nil {
		return fmt.Errorf("failed to grant privileges: %w", err)
	}

	query = fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO %s", conf.Database.Username)
	if _, err = conn.Exec(query); err != nil {
		return fmt.Errorf("failed to grant privileges: %w", err)
	}

	query = fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON FUNCTIONS TO %s", conf.Database.Username)
	if _, err = conn.Exec(query); err != nil {
		return fmt.Errorf("failed to grant privileges: %w", err)
	}

	slog.Info("createDatabaseAndUser DONE")

	return nil
}

func runMigrations(db *gorm.DB) error {
	// Ensure uuid-ossp is installed to this databse
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		return err
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Stash{},
		&models.ArtProject{},
		&models.Category{},
		&models.Tag{},
		&models.Revision{},
		&models.Collection{},
		&models.CollectionArtProject{},
		&models.Sale{},
		&models.StorageUsage{},
		&models.WebPage{},
	); err != nil {
		return err
	}
	return nil
}
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
	if err := createDatabaseAndUser(conf); err != nil {
		log.Fatalf("Failed to create database and user: %v", err)
	}

	// Database connection
	dbConnectionString := conf.Database.ConnectionString()
	slog.Info("database connection string", "dsn", dbConnectionString)
	db, err := gorm.Open(postgres.Open(conf.Database.ConnectionString()), &gorm.Config{
		DryRun: false,
	})
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
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
