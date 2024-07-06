package service

import (
	"context"
	"io"
	"os"

	"github.com/mirai-box/mirai-box/internal/model"
)

type PictureManagementService interface {
	CreatePictureAndRevision(ctx context.Context, fileData io.Reader, title, filename string) (*model.Picture, error)
	AddRevision(ctx context.Context, pictureID string, fileData io.Reader, comment, filename string) (*model.Revision, error)
	ListLatestRevisions(ctx context.Context) ([]model.Revision, error)
	ListAllPictures(ctx context.Context) ([]model.Picture, error)
	ListAllRevisions(ctx context.Context, pictureID string) ([]model.Revision, error)
}

type PictureRetrievalService interface {
	GetPictureByRevision(ctx context.Context, pictureID, revisionID string) (*os.File, *model.Picture, error)
	GetPictureByID(ctx context.Context, pictureID string) (*os.File, *model.Picture, error)
	GetSharedPicture(ctx context.Context, artID string) (*os.File, *model.Picture, error)
}

type UserService interface {
	Authenticate(ctx context.Context, username, password string) (*model.User, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
}

type GalleryService interface {
	CreateGallery(ctx context.Context, title string) (*model.Gallery, error)
	AddImageToGallery(ctx context.Context, galleryID, revisionID string) (*model.Gallery, error)
	PublishGallery(ctx context.Context, galleryID string) error
	GetGalleryByID(ctx context.Context, galleryID string) (*model.Gallery, error)
	ListGalleries(ctx context.Context) ([]model.Gallery, error)
	GetImagesByGalleryID(ctx context.Context, galleryID string) ([]model.Revision, error)
	GetMainGallery(ctx context.Context) ([]model.Revision, error)
}

type WebPageService interface {
	CreateWebPage(ctx context.Context, title, html string) (*model.WebPage, error)
	UpdateWebPage(ctx context.Context, id, title, html string) (*model.WebPage, error)
	DeleteWebPage(ctx context.Context, id string) error
	GetWebPage(ctx context.Context, id string) (*model.WebPage, error)
	ListWebPages(ctx context.Context) ([]model.WebPage, error)
	ListWebPagesByType(ctx context.Context, webPagesType string) ([]model.WebPage, error)
}
