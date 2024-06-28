package repository

import (
	"github.com/jmoiron/sqlx"

	"github.com/mirai-box/mirai-box/internal/model"
)

type SQLUserRepository struct {
	db *sqlx.DB
}

func NewSQLUserRepository(db *sqlx.DB) *SQLUserRepository {
	return &SQLUserRepository{db: db}
}

func (r *SQLUserRepository) FindByUsername(username string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, username, password, role FROM users WHERE username = $1`
	err := r.db.Get(user, query, username)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *SQLUserRepository) FindByID(id string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, username, password, role FROM users WHERE id = $1`
	err := r.db.Get(user, query, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}
