package service

import (
	"io"
	"os"

	"github.com/mirai-box/mirai-box/internal/model"
)

type PictureManagementService interface {
	CreatePictureAndRevision(fileData io.Reader, title, filename string) (*model.Picture, error)
	AddRevision(pictureID string, fileData io.Reader, comment, filename string) (*model.Revision, error)
	ListLatestRevisions() ([]model.Revision, error)
	ListAllPictures() ([]model.Picture, error)
	ListAllRevisions(pictureID string) ([]model.Revision, error)
}

type PictureRetrievalService interface {
	GetPictureByRevision(pictureID, revisionID string) (*os.File, *model.Picture, error)
	GetPictureByID(pictureID string) (*os.File, *model.Picture, error)
	GetSharedPicture(artID string) (*os.File, *model.Picture, error)
}

type UserService interface {
	Authenticate(username, password string) (*model.User, error)
	FindByID(id string) (*model.User, error)
}

type GalleryService interface {
	CreateGallery(title string) (*model.Gallery, error)
	AddImageToGallery(galleryID, revisionID string) (*model.Gallery, error)
	PublishGallery(galleryID string) error
	GetGalleryByID(galleryID string) (*model.Gallery, error)
	ListGalleries() ([]model.Gallery, error)
	GetImagesByGalleryID(galleryID string) ([]model.Revision, error)
}

type WebPageService interface {
	CreateWebPage(title, html string) (*model.WebPage, error)
	UpdateWebPage(id, title, html string) (*model.WebPage, error)
	DeleteWebPage(id string) error
	GetWebPage(id string) (*model.WebPage, error)
	ListWebPages() ([]model.WebPage, error)
}
