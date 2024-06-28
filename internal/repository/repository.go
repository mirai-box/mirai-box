package repository

import (
	"io"
	"os"

	"github.com/mirai-box/mirai-box/internal/model"
)

type PictureRepository interface {
	SavePicture(picture *model.Picture) error
	SaveRevision(revision *model.Revision) error
	GetMaxRevisionVersion(pictureID string) (int, error)
	ListLatestRevisions() ([]model.Revision, error)
	ListAllRevisions(pictureID string) ([]model.Revision, error)
	GetRevisionByID(revisionID string) (*model.Revision, error)
	GetRevisionByArtID(artID string) (*model.Revision, error)
	GetPictureByID(pictureID string) (*model.Picture, error)
	UpdateLatestRevision(pictureID, revisionID string) error
	ListAllPictures() ([]model.Picture, error)
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
}
