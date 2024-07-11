package models

import (
	"time"

	"github.com/google/uuid"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// ValidationErrorResponse represents a response for validation errors
type ValidationErrorResponse struct {
	Status  int               `json:"status"`
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors"`
}

// Pagination represents pagination information
type Pagination struct {
	CurrentPage  int `json:"current_page"`
	PerPage      int `json:"per_page"`
	TotalPages   int `json:"total_pages"`
	TotalRecords int `json:"total_records"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// ArtProjectResponse represents the response for an art project
type ArtProjectResponse struct {
	ID                  uuid.UUID  `json:"id"`
	Title               string     `json:"title"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	ContentType         string     `json:"content_type"`
	Filename            string     `json:"filename"`
	Public              bool       `json:"public"`
	LatestRevisionID    uuid.UUID  `json:"latest_revision_id"`
	PublishedRevisionID *uuid.UUID `json:"published_revision_id,omitempty"`
	Tags                []TagResponse `json:"tags"`
	StashID             uuid.UUID  `json:"stash_id"`
	UserID              uuid.UUID  `json:"user_id"`
}

// TagResponse represents the response for a tag
type TagResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// RevisionResponse represents the response for a revision
type RevisionResponse struct {
	ID           uuid.UUID `json:"id"`
	ArtID        string    `json:"art_id"`
	Version      int       `json:"version"`
	CreatedAt    time.Time `json:"created_at"`
	Comment      string    `json:"comment"`
	Size         int64     `json:"size"`
	ArtProjectID uuid.UUID `json:"art_project_id"`
	UserID       uuid.UUID `json:"user_id"`
}

// UserResponse represents the response for a user
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// StashResponse represents the response for a stash
type StashResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	ArtProjects uint64    `json:"art_projects"`
	Files       uint64    `json:"files"`
	UsedSpace   int64     `json:"used_space"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CollectionResponse represents the response for a collection
type CollectionResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CollectionArtProjectResponse represents the response for a collection art project
type CollectionArtProjectResponse struct {
	CollectionID uuid.UUID `json:"collection_id"`
	ArtProjectID uuid.UUID `json:"art_project_id"`
}

// CommentResponse represents the response for a comment
type CommentResponse struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	ArtProjectID uuid.UUID `json:"art_project_id"`
	Content      string    `json:"content"`
	CreatedAt    time.Time `json:"created_at"`
}

// SaleResponse represents the response for a sale
type SaleResponse struct {
	ID           uuid.UUID `json:"id"`
	ArtProjectID uuid.UUID `json:"art_project_id"`
	UserID       uuid.UUID `json:"user_id"`
	Price        float64   `json:"price"`
	SoldAt       time.Time `json:"sold_at"`
}

// StorageUsageResponse represents the response for storage usage
type StorageUsageResponse struct {
	UserID    uuid.UUID `json:"user_id"`
	UsedSpace int64     `json:"used_space"`
	Quota     int64     `json:"quota"`
}

// WebPageResponse represents the response for a web page
type WebPageResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Title     string    `json:"title"`
	Html      string    `json:"html"`
	PageType  string    `json:"page_type"`
	Public    bool      `json:"public"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ArtLinkResponse represents the response for an art link
type ArtLinkResponse struct {
	Token      string    `json:"token"`
	RevisionID uuid.UUID `json:"revision_id"`
	ExpiresAt  time.Time `json:"expires_at"`
	OneTime    bool      `json:"one_time"`
	Used       bool      `json:"used"`
}

// Request models

// CreateArtProjectRequest represents the request to create an art project
type CreateArtProjectRequest struct {
	Title string `json:"title"`
}

// AddRevisionRequest represents the request to add a revision
type AddRevisionRequest struct {
	Comment string `json:"comment"`
}