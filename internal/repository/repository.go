package repository

import (
	"context"
	"io"
	"os"

	"github.com/mirai-box/mirai-box/internal/model"
)

type PictureRepository interface {
	SaveRevision(revision *model.Revision) error
	GetMaxRevisionVersion(pictureID string) (int, error)
	ListLatestRevisions() ([]model.Revision, error)
	ListAllRevisions(pictureID string) ([]model.Revision, error)
	GetRevisionByID(revisionID string) (*model.Revision, error)
	GetRevisionByArtID(artID string) (*model.Revision, error)
	GetPictureByID(pictureID string) (*model.Picture, error)
	UpdateLatestRevision(pictureID, revisionID string) error
	ListAllPictures() ([]model.Picture, error)
	SavePictureAndRevision(picture *model.Picture, revision *model.Revision) error
}

type StorageRepository interface {
	SavePicture(fileData io.Reader, fileName string) error
	GetPicture(filePath string) (*os.File, error)
}

type UserRepository interface {
	FindByUsername(username string) (*model.User, error)
	FindByID(id string) (*model.User, error)
}

type GalleryRepository interface {
	CreateGallery(gallery *model.Gallery) error
	AddImageToGallery(galleryID, revisionID string) error
	PublishGallery(galleryID string) error
	GetGalleryByID(galleryID string) (*model.Gallery, error)
	ListGalleries() ([]model.Gallery, error)
	GetImagesByGalleryID(galleryID string) ([]model.Revision, error)
	GetMainGallery() (*model.Gallery, error)
}

// WebPageRepository defines the operations for managing web pages.
type WebPageRepository interface {
	CreateWebPage(ctx context.Context, wp *model.WebPage) error
	UpdateWebPage(ctx context.Context, wp *model.WebPage) error
	DeleteWebPage(ctx context.Context, id string) error
	GetWebPage(ctx context.Context, id string) (*model.WebPage, error)
	ListWebPages(ctx context.Context) ([]model.WebPage, error)
}
