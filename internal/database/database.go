package database

import (
	"log/slog"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/mirai-box/mirai-box/internal/config"
)

func NewConnection(cfg *config.Config) (*sqlx.DB, error) {
	db, err := setupDB(cfg.Database)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func setupDB(dbConf *config.DatabaseConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", dbConf.String())
	if err != nil {
		slog.Error("can't connect to database", "error", err, "dataSourceName", dbConf.String())
		return nil, err
	}

	if err := db.Ping(); err != nil {
		slog.Error("can't ping database", "error", err, "dataSourceName", dbConf.String())
		return nil, err
	}

	return db, nil
}
