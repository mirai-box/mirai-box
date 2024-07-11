package repos

import (
	"context"
	"io"
	"os"

	"github.com/google/uuid"
	"github.com/mirai-box/mirai-box/internal/models"
)

// StorageUsageRepositoryInterface defines the contract for storage usage-related database operations
type StorageUsageRepositoryInterface interface {
	Create(ctx context.Context, storageUsage *models.StorageUsage) error
	FindByUserID(ctx context.Context, userID string) (*models.StorageUsage, error)
	Update(ctx context.Context, storageUsage *models.StorageUsage) error
	Delete(ctx context.Context, userID string) error
}

// ArtProjectRepositoryInterface defines the contract for art project-related database operations
type ArtProjectRepositoryInterface interface {
	Create(ctx context.Context, artProject *models.ArtProject) error
	FindByID(ctx context.Context, id string) (*models.ArtProject, error)
	SaveArtProjectAndRevision(ctx context.Context, artProject *models.ArtProject, revision *models.Revision) error
	SaveRevision(ctx context.Context, revision *models.Revision) error
	UpdateLatestRevision(ctx context.Context, artProjectID string, revisionID uuid.UUID) error
	ListLatestRevisions(ctx context.Context, userID string) ([]models.Revision, error)
	ListAllArtProjects(ctx context.Context, userID string) ([]models.ArtProject, error)
	ListAllRevisions(ctx context.Context, artProjectID string) ([]models.Revision, error)
	GetMaxRevisionVersion(ctx context.Context, artProjectID string) (int, error)
	GetRevisionByID(ctx context.Context, id string) (*models.Revision, error)
	GetArtProjectByID(ctx context.Context, id string) (*models.ArtProject, error)
	FindByStashID(ctx context.Context, stashID string) ([]models.ArtProject, error)
	CreateArtLink(ctx context.Context, artLink *models.ArtLink) error
	UpdateArtLink(ctx context.Context, artLink *models.ArtLink) error
	GetArtLinkByToken(ctx context.Context, token string) (*models.ArtLink, error)
	FindByUserID(ctx context.Context, userID string) ([]models.ArtProject, error)
}

// StorageRepositoryInterface defines the contract for storage-related operations
type StorageRepositoryInterface interface {
	SaveRevision(ctx context.Context, fileData io.Reader, userID, artProjectID string, version int) (string, os.FileInfo, error)
	GetRevision(ctx context.Context, userID, artProjectID string, version int) (*os.File, error)
}

// CollectionArtProjectRepositoryInterface defines the contract for collection art project-related database operations
type CollectionArtProjectRepositoryInterface interface {
	Create(ctx context.Context, collectionArtProject *models.CollectionArtProject) error
	FindByCollectionID(ctx context.Context, collectionID string) ([]models.CollectionArtProject, error)
	FindByArtProjectID(ctx context.Context, artProjectID string) ([]models.CollectionArtProject, error)
	Delete(ctx context.Context, collectionID string, artProjectID string) error
}

// CollectionRepositoryInterface defines the contract for collection-related database operations
type CollectionRepositoryInterface interface {
	Create(ctx context.Context, collection *models.Collection) error
	FindByID(ctx context.Context, id string) (*models.Collection, error)
	FindByUserID(ctx context.Context, userID string) ([]models.Collection, error)
	Update(ctx context.Context, collection *models.Collection) error
	Delete(ctx context.Context, id string) error
}

// RevisionRepositoryInterface defines the contract for revision-related database operations
type RevisionRepositoryInterface interface {
	Create(ctx context.Context, revision *models.Revision) error
	FindByID(ctx context.Context, id string) (*models.Revision, error)
	FindByArtProjectID(ctx context.Context, artProjectID string) ([]models.Revision, error)
	Update(ctx context.Context, revision *models.Revision) error
	Delete(ctx context.Context, id string) error
}

// SaleRepositoryInterface defines the contract for sale-related database operations
type SaleRepositoryInterface interface {
	Create(ctx context.Context, sale *models.Sale) error
	FindByID(ctx context.Context, id string) (*models.Sale, error)
	FindByUserID(ctx context.Context, userID string) ([]models.Sale, error)
	FindByArtProjectID(ctx context.Context, artProjectID string) ([]models.Sale, error)
	Update(ctx context.Context, sale *models.Sale) error
	Delete(ctx context.Context, id string) error
}

// StashRepositoryInterface defines the contract for stash-related database operations
type StashRepositoryInterface interface {
	Create(ctx context.Context, stash *models.Stash) error
	FindByID(ctx context.Context, id string) (*models.Stash, error)
	FindByUserID(ctx context.Context, userID string) (*models.Stash, error)
	Update(ctx context.Context, stash *models.Stash) error
	Delete(ctx context.Context, id string) error
}

// WebPageRepositoryInterface defines the contract for web page-related database operations
type WebPageRepositoryInterface interface {
	Create(ctx context.Context, webPage *models.WebPage) error
	Update(ctx context.Context, webPage *models.WebPage) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id string) (*models.WebPage, error)
	FindAll(ctx context.Context) ([]models.WebPage, error)
	FindByType(ctx context.Context, pageType string) ([]models.WebPage, error)
	FindByUserID(ctx context.Context, userID string) ([]models.WebPage, error)
}

// UserRepositoryInterface defines the contract for user-related database operations
type UserRepositoryInterface interface {
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id string) (*models.User, error)
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error
}
