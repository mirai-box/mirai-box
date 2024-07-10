package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

const (
	localStage        = "local"
	defaultDebugLevel = "info"
	defaultDBPort     = "5432"
	defaultDBHost     = "127.0.0.1"
	defaultDBName     = "mirai_box_db"
	defaultDBUser     = "mirai_box_user"
	defaultAppStage   = localStage
	defaultPort       = "8080"
)

type Config struct {
	LogLevel    slog.Level
	Stage       string
	Port        string
	Database    *DatabaseConfig
	StorageRoot string
	ProjectRoot string
	SessionKey  string
}

type DatabaseConfig struct {
	Host             string
	Port             string
	Database         string
	Username         string
	Password         string
	PostgresPassword string
	SSLMode          string
}

func GetApplicationConfig() (*Config, error) {
	projectRoot := getEnv("PROJECT_ROOT", getCurrentDir())
	storageRoot := getEnv("STORAGE_ROOT", filepath.Join(projectRoot, "storage"))

	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		return nil, fmt.Errorf("SESSION_KEY environment variable is not set")
	}

	return &Config{
		Stage:       getEnv("APP_ENV", defaultAppStage),
		Port:        getEnv("PORT", defaultPort),
		StorageRoot: storageRoot,
		ProjectRoot: projectRoot,
		SessionKey:  sessionKey,
		LogLevel:    parseLogLevel(getEnv("LOG_LEVEL", defaultDebugLevel)),
		Database:    GetDatabaseConfig(),
	}, nil
}

func GetDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:             getEnv("DB_HOST", defaultDBHost),
		Port:             getEnv("DB_PORT", defaultDBPort),
		Username:         getEnv("DB_USER", defaultDBUser),
		Password:         os.Getenv("DB_PASSWORD"),
		PostgresPassword: os.Getenv("DB_POSTGRES_PASSWORD"),
		Database:         getEnv("DB_NAME", defaultDBName),
		SSLMode:          getEnv("DB_SSLMODE", "disable"),
	}
}

func (conf *Config) IsLocal() bool {
	return conf.Stage == localStage
}

func (dbConfig *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.Username, dbConfig.Password, dbConfig.Database, dbConfig.SSLMode)
}

func (dbConfig *DatabaseConfig) ConnectionPostgres() string {
	return fmt.Sprintf("host=%s port=%s user=postgres dbname=postgres password=%s sslmode=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.PostgresPassword, dbConfig.SSLMode)
}

func parseLogLevel(s string) slog.Level {
	var level slog.Level
	if err := level.UnmarshalText([]byte(s)); err != nil {
		return slog.LevelInfo
	}
	return level
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		slog.Error("Failed to get current directory", "error", err)
		return ""
	}
	return dir
}
