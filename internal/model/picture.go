package model

import (
	"time"
)

// Picture represents the image entity with a UUID.
type Picture struct {
	ID                  string    `db:"id" json:"id"`
	Title               string    `db:"title" json:"title"`
	ContentType         string    `db:"content_type" json:"content_type"`
	Filename            string    `db:"filename" json:"filename"`
	Tags                []string  `db:"tags" json:"tags"`
	CreatedAt           time.Time `db:"created_at" json:"created_at"`
	LatestRevisionID    string    `db:"latest_revision_id" json:"latest_revision_id"`
	PublishedRevisionID string    `db:"published_revision_id" json:"published_revision_id"`
	LatestComment       string    `json:"latest_comment"`
	ArtID               string    `json:"art_id"`
}
