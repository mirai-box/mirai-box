package model

import (
	"time"
)

// Revision represents a specific version of a picture.
type Revision struct {
	ID        string    `db:"id"         json:"id"`
	PictureID string    `db:"picture_id" json:"picture_id"`
	ArtID     string    `db:"art_id"     json:"art_id"`
	Version   int       `db:"version"    json:"version"`
	FilePath  string    `db:"file_path"  json:"-"`
	Comment   string    `db:"comment"    json:"comment"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
