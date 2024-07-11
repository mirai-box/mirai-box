package service

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/mirai-box/mirai-box/internal/models"
)

// ArtProjectRetrievalServiceInterface defines the interface for retrieving art projects and revisions.
type ArtProjectRetrievalServiceInterface interface {
	GetRevisionByArtID(ctx context.Context, artID string) (*models.Revision, error)
	CreateArtLink(ctx context.Context, revisionID uuid.UUID, duration time.Duration, oneTime bool, unlimited bool) (string, error)
	GetArtByToken(ctx context.Context, token string) (*models.Revision, error)
	GetArtProjectByRevision(ctx context.Context, userID, artProjectID, revisionID string) (*os.File, *models.ArtProject, error)
	GetArtProjectByID(ctx context.Context, userID, artProjectID string) (*os.File, *models.ArtProject, error)
}

// ArtProjectManagementServiceInterface defines the interface for managing art projects and revisions.
type ArtProjectManagementServiceInterface interface {
	CreateArtProjectAndRevision(ctx context.Context, userID string, fileData io.Reader, title, filename string) (*models.ArtProject, error)
	AddRevision(ctx context.Context, userID, artProjectID string, fileData io.Reader, comment, filename string) (*models.Revision, error)
	ListLatestRevisions(ctx context.Context, userID string) ([]models.Revision, error)
	ListAllArtProjects(ctx context.Context, userID string) ([]models.ArtProject, error)
	ListAllRevisions(ctx context.Context, artProjectID string) ([]models.Revision, error)
}

// ArtProjectServiceInterface defines the interface for managing art projects.
type ArtProjectServiceInterface interface {
	FindByUserID(ctx context.Context, userID string) ([]models.ArtProject, error)
	CreateArtProject(ctx context.Context, stashID, title string) (*models.ArtProject, error)
	FindByID(ctx context.Context, id string) (*models.ArtProject, error)
	FindByStashID(ctx context.Context, stashID string) ([]models.ArtProject, error)
}

// CollectionArtProjectServiceInterface defines the interface for managing collection art projects.
type CollectionArtProjectServiceInterface interface {
	AddArtProjectToCollection(ctx context.Context, collectionID, artProjectID string) (*models.CollectionArtProject, error)
	FindByCollectionID(ctx context.Context, collectionID string) ([]models.CollectionArtProject, error)
	FindByArtProjectID(ctx context.Context, artProjectID string) ([]models.CollectionArtProject, error)
}

// CollectionServiceInterface defines the interface for managing collections.
type CollectionServiceInterface interface {
	CreateCollection(ctx context.Context, userID, title string) (*models.Collection, error)
	FindByID(ctx context.Context, id string) (*models.Collection, error)
	FindByUserID(ctx context.Context, userID string) ([]models.Collection, error)
}

// RevisionServiceInterface defines the interface for managing revisions.
type RevisionServiceInterface interface {
	CreateRevision(ctx context.Context, artProjectID, filePath, comment string, version int, size int64) (*models.Revision, error)
	FindByID(ctx context.Context, id string) (*models.Revision, error)
	FindByArtProjectID(ctx context.Context, artProjectID string) ([]models.Revision, error)
}

// SaleServiceInterface defines the contract for sale-related operations
type SaleServiceInterface interface {
	CreateSale(ctx context.Context, artProjectID, userID string, price float64) (*models.Sale, error)
	FindByID(ctx context.Context, id string) (*models.Sale, error)
	FindByUserID(ctx context.Context, userID string) ([]models.Sale, error)
	FindByArtProjectID(ctx context.Context, artProjectID string) ([]models.Sale, error)
}

// WebPageServiceInterface defines the contract for web page-related operations
type WebPageServiceInterface interface {
	CreateWebPage(ctx context.Context, webPage *models.WebPage) (*models.WebPage, error)
	UpdateWebPage(ctx context.Context, webPage *models.WebPage) (*models.WebPage, error)
	DeleteWebPage(ctx context.Context, id string) error
	GetWebPage(ctx context.Context, id string) (*models.WebPage, error)
	ListWebPages(ctx context.Context) ([]models.WebPage, error)
	ListUserWebPages(ctx context.Context, userID string) ([]models.WebPage, error)
	ListWebPagesByType(ctx context.Context, pageType string) ([]models.WebPage, error)
}

// UserServiceInterface defines the contract for user-related operations
type UserServiceInterface interface {
	Authenticate(ctx context.Context, username, password string) (*models.User, error)
	FindByID(ctx context.Context, id string) (*models.User, error)
	CreateUser(ctx context.Context, username, password, role string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) (*models.User, error)
	DeleteUser(ctx context.Context, id string) error
}

// StashServiceInterface defines the contract for stash-related operations
type StashServiceInterface interface {
	CreateStash(ctx context.Context, userID string) (*models.Stash, error)
	FindByID(ctx context.Context, id string) (*models.Stash, error)
	FindByUserID(ctx context.Context, userID string) (*models.Stash, error)
}

// StorageUsageServiceInterface defines the contract for storage usage-related operations
type StorageUsageServiceInterface interface {
	CreateStorageUsage(ctx context.Context, userID string, quota int64) (*models.StorageUsage, error)
	FindByUserID(ctx context.Context, userID string) (*models.StorageUsage, error)
	UpdateStorageUsage(ctx context.Context, userID string, usedSpace int64) (*models.StorageUsage, error)
}
