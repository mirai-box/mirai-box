package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/mirai-box/mirai-box/internal/model"
)

type SQLGalleryRepository struct {
	db *sqlx.DB
}

func NewSQLGalleryRepository(db *sqlx.DB) GalleryRepository {
	return &SQLGalleryRepository{db: db}
}

func (r *SQLGalleryRepository) CreateGallery(ctx context.Context, gallery *model.Gallery) error {
	query := `
INSERT INTO galleries (id, title, created_at, published) 
VALUES (:id, :title, :created_at, :published)`

	_, err := r.db.NamedExec(query, gallery)
	return err
}

func (r *SQLGalleryRepository) AddImageToGallery(ctx context.Context, galleryID, revisionID string) error {
	query := `INSERT INTO gallery_images (gallery_id, revision_id) VALUES ($1, $2)`
	_, err := r.db.Exec(query, galleryID, revisionID)
	return err
}

func (r *SQLGalleryRepository) PublishGallery(ctx context.Context, galleryID string) error {
	query := `UPDATE galleries SET published = TRUE WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, galleryID)
	return err
}

func (r *SQLGalleryRepository) GetGalleryByID(ctx context.Context, galleryID string) (*model.Gallery, error) {
	var gallery model.Gallery
	query := `SELECT id, title, created_at, published FROM galleries WHERE id = $1`
	err := r.db.GetContext(ctx, &gallery, query, galleryID)
	if err != nil {
		return nil, err
	}
	return &gallery, nil
}

func (r *SQLGalleryRepository) ListGalleries(ctx context.Context) ([]model.Gallery, error) {
	var galleries []model.Gallery
	query := `SELECT id, title, created_at, published FROM galleries`
	err := r.db.Select(&galleries, query)
	if err != nil {
		return nil, err
	}
	return galleries, nil
}

func (r *SQLGalleryRepository) GetImagesByGalleryID(ctx context.Context, galleryID string) ([]model.Revision, error) {
	var revisions []model.Revision
	query := `
SELECT r.id, r.file_path, r.art_id, r.version, r.picture_id, r.comment, r.created_at
FROM revisions r
JOIN gallery_images gi ON r.id = gi.revision_id
WHERE gi.gallery_id = $1`

	err := r.db.SelectContext(ctx, &revisions, query, galleryID)
	if err != nil {
		return nil, err
	}
	return revisions, nil
}

// GetMainGallery retrieves the main gallery
func (r *SQLGalleryRepository) GetGalleryByTitle(ctx context.Context, title string) (*model.Gallery, error) {
	var gallery model.Gallery
	query := "SELECT id, title, type FROM galleries WHERE title = $1"

	err := r.db.QueryRowContext(ctx, query, title).
		Scan(&gallery.ID, &gallery.Title, &gallery.GalleryType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("gallery with title %q not found", title)
		}
		return nil, err
	}
	return &gallery, nil
}

func (r *SQLGalleryRepository) ListGallerisByType(ctx context.Context, galleryType string) ([]model.Gallery, error) {
	var galleries []model.Gallery
	query := "SELECT id, title, type FROM galleries WHERE type = $1"

	err := r.db.SelectContext(ctx, &galleries, query, galleryType)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("gallery with galleryType %q not found", galleryType)
		}
		return nil, err
	}
	return galleries, nil
}
