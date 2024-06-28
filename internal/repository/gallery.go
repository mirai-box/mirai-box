package repository

import (
	"github.com/jmoiron/sqlx"

	"github.com/mirai-box/mirai-box/internal/model"
)

type SQLGalleryRepository struct {
	db *sqlx.DB
}

func NewSQLGalleryRepository(db *sqlx.DB) GalleryRepository {
	return &SQLGalleryRepository{db: db}
}

func (r *SQLGalleryRepository) CreateGallery(gallery *model.Gallery) error {
	query := `
INSERT INTO galleries (id, title, created_at, published) 
VALUES (:id, :title, :created_at, :published)`

	_, err := r.db.NamedExec(query, gallery)
	return err
}

func (r *SQLGalleryRepository) AddImageToGallery(galleryID, revisionID string) error {
	query := `INSERT INTO gallery_images (gallery_id, revision_id) VALUES ($1, $2)`
	_, err := r.db.Exec(query, galleryID, revisionID)
	return err
}

func (r *SQLGalleryRepository) PublishGallery(galleryID string) error {
	query := `UPDATE galleries SET published = TRUE WHERE id = $1`
	_, err := r.db.Exec(query, galleryID)
	return err
}

func (r *SQLGalleryRepository) GetGalleryByID(galleryID string) (*model.Gallery, error) {
	var gallery model.Gallery
	query := `SELECT id, title, created_at, published FROM galleries WHERE id = $1`
	err := r.db.Get(&gallery, query, galleryID)
	if err != nil {
		return nil, err
	}
	return &gallery, nil
}

func (r *SQLGalleryRepository) ListGalleries() ([]model.Gallery, error) {
	var galleries []model.Gallery
	query := `SELECT id, title, created_at, published FROM galleries`
	err := r.db.Select(&galleries, query)
	if err != nil {
		return nil, err
	}
	return galleries, nil
}

func (r *SQLGalleryRepository) GetImagesByGalleryID(galleryID string) ([]model.Revision, error) {
	var revisions []model.Revision
	query := `
SELECT r.id, r.file_path, r.art_id, r.version, r.picture_id, r.comment, r.created_at
FROM revisions r
JOIN gallery_images gi ON r.id = gi.revision_id
WHERE gi.gallery_id = $1`
	err := r.db.Select(&revisions, query, galleryID)
	if err != nil {
		return nil, err
	}
	return revisions, nil
}
