package model 

import "time"

// WebPage represents a single web page.
type WebPage struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    HTML      string    `json:"html"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
