package repository

import (
	"database/sql"
	"log/slog"

	"github.com/jmoiron/sqlx"
	"github.com/mirai-box/mirai-box/internal/model"
)

type sqlPictureRepository struct {
	db *sqlx.DB
}

func NewPictureRepository(db *sqlx.DB) PictureRepository {
	return &sqlPictureRepository{
		db: db,
	}
}

func (r *sqlPictureRepository) SaveRevision(revision *model.Revision) error {
	query := `INSERT INTO revisions (
		id, 
		picture_id, 
		version, 
		file_path, 
		comment,
		art_id,
		created_at) 
	VALUES (:id, :picture_id, :version, :file_path, :comment, :art_id, :created_at)`
	_, err := r.db.NamedExec(query, revision)
	return err
}

func (r *sqlPictureRepository) SavePictureAndRevision(picture *model.Picture, revision *model.Revision) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert revision first
	queryRev := `INSERT INTO revisions (
		id, 
		picture_id, 
		version, 
		file_path, 
		comment,
		art_id,
		created_at) 
	VALUES (:id, :picture_id, :version, :file_path, :comment, :art_id, :created_at)`
	_, err = tx.NamedExec(queryRev, revision)
	if err != nil {
		return err
	}

	// Now insert picture with the revision ID
	picture.LatestRevisionID = revision.ID
	queryPic := `INSERT INTO pictures
		(id, title, filename, content_type, latest_revision_id, created_at) 
		VALUES (:id, :title, :filename, :content_type, :latest_revision_id, :created_at)`
	_, err = tx.NamedExec(queryPic, picture)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *sqlPictureRepository) GetMaxRevisionVersion(pictureID string) (int, error) {
	var maxVersion int
	query := "SELECT MAX(version) FROM revisions WHERE picture_id = $1"
	err := r.db.Get(&maxVersion, query, pictureID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return maxVersion, nil
}

func (r *sqlPictureRepository) ListLatestRevisions() ([]model.Revision, error) {
	const query = `
SELECT id, picture_id, version, comment, art_id, created_at
FROM revisions
WHERE id IN (
    SELECT MAX(id)
    FROM revisions
    GROUP BY picture_id
);`

	var revisions []model.Revision
	err := r.db.Select(&revisions, query)
	if err != nil {
		return nil, err
	}

	return revisions, nil
}

func (r *sqlPictureRepository) GetRevisionByID(revisionID string) (*model.Revision, error) {
	var rev model.Revision
	query := `SELECT id, picture_id, version, file_path, comment, art_id, created_at
              FROM revisions WHERE id = $1`

	err := r.db.Get(&rev, query, revisionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &rev, nil
}

func (r *sqlPictureRepository) GetRevisionByArtID(artID string) (*model.Revision, error) {
	slog.Debug("db: get revision by shared art id", "artID", artID)

	var rev model.Revision
	query := `
SELECT 
	id, 
	picture_id, 
	version, 
	file_path, 
	comment, 
	art_id, 
	created_at
FROM revisions WHERE art_id = $1`

	err := r.db.Get(&rev, query, artID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &rev, nil
}

func (r *sqlPictureRepository) GetPictureByID(pictureID string) (*model.Picture, error) {
	var picture model.Picture
	query := `
SELECT 
	id, 
	title, 
	content_type, 
	filename, 
	latest_revision_id, 
	created_at 
FROM pictures WHERE id = $1`

	err := r.db.Get(&picture, query, pictureID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &picture, nil
}

func (r *sqlPictureRepository) UpdateLatestRevision(pictureID, revisionID string) error {
	query := `UPDATE pictures SET latest_revision_id = $1 WHERE id = $2`
	_, err := r.db.Exec(query, revisionID, pictureID)
	return err
}

func (r *sqlPictureRepository) ListAllRevisions(pictureID string) ([]model.Revision, error) {
	query := `
SELECT id, picture_id, version, file_path, created_at, comment, art_id
FROM revisions
WHERE picture_id = $1
ORDER BY version DESC;`

	var revisions []model.Revision
	err := r.db.Select(&revisions, query, pictureID)
	if err != nil {
		return nil, err
	}

	return revisions, nil
}

func (r *sqlPictureRepository) ListAllPictures() ([]model.Picture, error) {
	query := `
SELECT 
    p.id, 
    p.title, 
    p.content_type, 
    p.filename, 
    p.created_at,
	p.latest_revision_id,
	r.art_id,
    r.comment as latest_comment
FROM pictures p
LEFT JOIN revisions r ON p.latest_revision_id = r.id;`

	var pictures []model.Picture
	rows, err := r.db.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p model.Picture
		var latestComment sql.NullString
		if err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.ContentType,
			&p.Filename,
			&p.CreatedAt,
			&p.LatestRevisionID,
			&p.ArtID,
			&latestComment,
		); err != nil {
			return nil, err
		}
		p.LatestComment = latestComment.String
		pictures = append(pictures, p)
	}

	return pictures, rows.Err()
}
