package config

import (
	"fmt"
	"log/slog"
	"os"
	"path"
)

const (
	localStage        = "local"
	defaultDebugLevel = "debug"
	defualtDBPort     = "5432"
	defaultDBHost     = "127.0.0.1"
	defaultAppStage   = localStage
)

type Config struct {
	LogLevel    slog.Level
	Stage       string
	Port        string
	IsCGI       bool
	Database    *DatabaseConfig
	StorageRoot string
	ProjectRoot string
	SessionKey  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
	SSLMode  string
}

func GetApplicationConfig() (*Config, error) {
	// Detect if running in a CGI environment
	cgi := os.Getenv("GATEWAY_INTERFACE") != ""

	// Configure port
	port := os.Getenv("PORT")
	if port == "" {
		if cgi {
			port = "" // In CGI mode, the port is not used.
		} else {
			port = "8080" // default HTTP port
		}
	}

	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		// Get the current working directory
		cwd, _ := os.Getwd()
		projectRoot = cwd
	}

	storageRoot := os.Getenv("STORAGE_ROOT")
	if storageRoot == "" {
		storageRoot = path.Join(projectRoot, "storage")
	}

	// Get the session key from the environment variable
	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		panic("SESSION_KEY environment variable is not set")
	}

	return &Config{
		Stage:       getEnv("APP_ENV", defaultAppStage),
		Port:        port,
		IsCGI:       cgi,
		StorageRoot: storageRoot,
		ProjectRoot: projectRoot,
		SessionKey:  sessionKey,
		LogLevel:    parseLogLevel(getEnv("DEBUG_LEVEL", defaultDebugLevel)),
		Database: &DatabaseConfig{
			Host:     getEnv("DB_HOST", defaultDBHost),
			Port:     getEnv("DB_PORT", defualtDBPort),
			Username: getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Database: getEnv("DB_NAME", "picture_db"),
			SSLMode:  "disable",
		},
	}, nil
}

func (conf *Config) IsLocal() bool {
	return conf.Stage == localStage
}

func (dbConfig *DatabaseConfig) String() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.Username, dbConfig.Password, dbConfig.Database, dbConfig.SSLMode)
}

func parseLogLevel(s string) slog.Level {
	var level slog.Level
	if err := level.UnmarshalText([]byte(s)); err != nil {
		return slog.LevelInfo
	}
	return level
}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}
