package repository

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/mirai-box/mirai-box/internal/model"
)

// SQLWebPageRepository is a SQL-based implementation of WebPageRepository.
type SQLWebPageRepository struct {
	db *sqlx.DB
}

// NewSQLWebPageRepository creates a new SQLWebPageRepository.
func NewWebPageRepository(db *sqlx.DB) WebPageRepository {
	return &SQLWebPageRepository{db: db}
}

// CreateWebPage creates a new web page.
func (r *SQLWebPageRepository) CreateWebPage(ctx context.Context, wp *model.WebPage) error {
	query := `
		INSERT INTO web_pages (id, title, html, page_type, created_at, updated_at) 
		VALUES (:id, :title, :html, :page_type, :created_at, :updated_at)
	`
	_, err := r.db.NamedExecContext(ctx, query, wp)
	return err
}

// UpdateWebPage updates an existing web page.
func (r *SQLWebPageRepository) UpdateWebPage(ctx context.Context, wp *model.WebPage) error {
	query := `
		UPDATE web_pages 
		SET title = :title, html = :html, page_type = :page_type, updated_at = :updated_at
		WHERE id = :id
	`
	_, err := r.db.NamedExecContext(ctx, query, wp)
	return err
}

// DeleteWebPage deletes a web page by ID.
func (r *SQLWebPageRepository) DeleteWebPage(ctx context.Context, id string) error {
	query := `DELETE FROM web_pages WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GetWebPage retrieves a web page by ID.
func (r *SQLWebPageRepository) GetWebPage(ctx context.Context, id string) (*model.WebPage, error) {
	var wp model.WebPage
	query := `
		SELECT id, title, html, page_type, created_at, updated_at
		FROM web_pages
		WHERE id = $1
	`
	err := r.db.GetContext(ctx, &wp, query, id)
	if err != nil {
		return nil, err
	}
	return &wp, nil
}

// ListWebPages retrieves all web pages.
func (r *SQLWebPageRepository) ListWebPages(ctx context.Context) ([]model.WebPage, error) {
	var webPages []model.WebPage
	query := `
		SELECT id, title, html, page_type, created_at, updated_at
		FROM web_pages
	`
	err := r.db.SelectContext(ctx, &webPages, query)
	if err != nil {
		return nil, err
	}
	return webPages, nil
}

// pages retrieves all web pages by given type.
func (r *SQLWebPageRepository) ListWebPagesByType(ctx context.Context, webPageType string) ([]model.WebPage, error) {
	var webPages []model.WebPage
	query := `
		SELECT id, title, html, page_type, created_at, updated_at
		FROM web_pages
		WHERE page_type = $1
	`
	err := r.db.SelectContext(ctx, &webPages, query, webPageType)
	if err != nil {
		return nil, err
	}
	return webPages, nil
}
