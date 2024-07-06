package model

import "time"

// WebPage represents a single web page.
type WebPage struct {
	ID        string    `json:"id"         db:"id"`
	Title     string    `json:"title"      db:"title"`
	HTML      string    `json:"html"       db:"html"`
	PageType  string    `json:"page_type"  db:"page_type"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
