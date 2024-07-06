package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/mirai-box/mirai-box/internal/model"
	"github.com/mirai-box/mirai-box/internal/repository"
)

// galleryService implements business logic for galleries.
type galleryService struct {
	galleryRepo repository.GalleryRepository
}

// NewGalleryService creates a new gallery service.
func NewGalleryService(galleryRepo repository.GalleryRepository) *galleryService {
	return &galleryService{galleryRepo: galleryRepo}
}

// CreateGallery creates a new gallery with the given title.
func (s *galleryService) CreateGallery(ctx context.Context, title string) (*model.Gallery, error) {
	gallery := &model.Gallery{
		ID:        uuid.NewString(),
		Title:     title,
		CreatedAt: time.Now(),
		Published: false,
	}

	err := s.galleryRepo.CreateGallery(ctx, gallery)
	if err != nil {
		return nil, err
	}

	return gallery, nil
}

// AddImageToGallery adds an image (identified by its revision ID) to a gallery.
func (s *galleryService) AddImageToGallery(ctx context.Context, galleryID, revisionID string) (*model.Gallery, error) {
	// Validate that a revision ID is provided.
	if revisionID == "" {
		return nil, fmt.Errorf("revisionID can't be empty")
	}

	// Add the image to the gallery in the repository.
	err := s.galleryRepo.AddImageToGallery(ctx, galleryID, revisionID)
	if err != nil {
		return nil, err
	}

	// Return the updated gallery.
	return s.galleryRepo.GetGalleryByID(ctx, galleryID)
}

func (s *galleryService) PublishGallery(ctx context.Context, galleryID string) error {
	return s.galleryRepo.PublishGallery(ctx, galleryID)
}

func (s *galleryService) GetGalleryByID(ctx context.Context, galleryID string) (*model.Gallery, error) {
	return s.galleryRepo.GetGalleryByID(ctx, galleryID)
}

func (s *galleryService) ListGalleries(ctx context.Context) ([]model.Gallery, error) {
	galleries, err := s.galleryRepo.ListGalleries(ctx)
	if err != nil {
		return nil, err
	}

	if len(galleries) == 0 {
		return []model.Gallery{}, nil
	}

	return galleries, nil
}

func (s *galleryService) GetImagesByGalleryID(ctx context.Context, galleryID string) ([]model.Revision, error) {
	revisions, err := s.galleryRepo.GetImagesByGalleryID(ctx, galleryID)
	if err != nil {
		return nil, err
	}

	if len(revisions) == 0 {
		return []model.Revision{}, nil
	}

	return revisions, nil
}

// GetMainGallery retrieves the main gallery
func (s *galleryService) GetMainGallery(ctx context.Context) ([]model.Revision, error) {
	gallery, err := s.galleryRepo.GetGalleryByTitle(context.Background(), "Main")
	if err != nil {
		return nil, err
	}

	return s.GetImagesByGalleryID(ctx, gallery.ID)
}
