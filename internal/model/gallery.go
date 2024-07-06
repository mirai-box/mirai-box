package model

import "time"

// Gallery represents a collection of selected images.
type Gallery struct {
	ID          string    `db:"id"           json:"id"`
	Title       string    `db:"title"        json:"title"`
	GalleryType string    `db:"gallery_type" json:"gallery_type"`
	CreatedAt   time.Time `db:"created_at"   json:"created_at"`
	Published   bool      `db:"published"    json:"published"`
}

// GalleryImage represents the association between a gallery and a picture.
type GalleryImage struct {
	GalleryID string `db:"gallery_id" json:"gallery_id"`
	PictureID string `db:"picture_id" json:"picture_id"`
}
