package database

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"gorm.io/gorm"

	"github.com/mirai-box/mirai-box/internal/config"
	"github.com/mirai-box/mirai-box/internal/models"
)

// CreateDatabaseAndUser creates the database and user if they don't exist
func CreateDatabaseAndUser(conf *config.Config) error {
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
	queries := []string{
		fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE %s TO %s", conf.Database.Database, conf.Database.Username),
		fmt.Sprintf("GRANT ALL PRIVILEGES ON SCHEMA public TO %s", conf.Database.Username),
		fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO %s", conf.Database.Username),
		fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO %s", conf.Database.Username),
		fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON FUNCTIONS TO %s", conf.Database.Username),
	}

	for _, query := range queries {
		if _, err = conn.Exec(query); err != nil {
			return fmt.Errorf("failed to grant privileges: %w", err)
		}
	}

	return nil
}

// RunMigrations runs the database migrations
func RunMigrations(db *gorm.DB) error {
	// Ensure uuid-ossp is installed to this database
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		return err
	}

	return db.AutoMigrate(
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
		&models.ArtLink{},
	)
}
