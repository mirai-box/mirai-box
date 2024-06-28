package service

import (
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
func (s *galleryService) CreateGallery(title string) (*model.Gallery, error) {
	gallery := &model.Gallery{
		ID:        uuid.NewString(),
		Title:     title,
		CreatedAt: time.Now(),
		Published: false,
	}
	err := s.galleryRepo.CreateGallery(gallery)
	if err != nil {
		return nil, err
	}
	return gallery, nil
}

// AddImageToGallery adds an image (identified by its revision ID) to a gallery.
func (s *galleryService) AddImageToGallery(galleryID, revisionID string) (*model.Gallery, error) {
	// Validate that a revision ID is provided.
	if revisionID == "" {
		return nil, fmt.Errorf("revisionID can't be empty")
	}

	// Add the image to the gallery in the repository.
	err := s.galleryRepo.AddImageToGallery(galleryID, revisionID)
	if err != nil {
		return nil, err
	}

	// Return the updated gallery.
	return s.galleryRepo.GetGalleryByID(galleryID)
}

func (s *galleryService) PublishGallery(galleryID string) error {
	return s.galleryRepo.PublishGallery(galleryID)
}

func (s *galleryService) GetGalleryByID(galleryID string) (*model.Gallery, error) {
	return s.galleryRepo.GetGalleryByID(galleryID)
}

func (s *galleryService) ListGalleries() ([]model.Gallery, error) {
	galleries, err := s.galleryRepo.ListGalleries()
	if err != nil {
		return nil, err
	}

	if len(galleries) == 0 {
		return []model.Gallery{}, nil
	}

	return galleries, nil
}

func (s *galleryService) GetImagesByGalleryID(galleryID string) ([]model.Revision, error) {
	revisions, err := s.galleryRepo.GetImagesByGalleryID(galleryID)
	if err != nil {
		return nil, err
	}

	if len(revisions) == 0 {
		return []model.Revision{}, nil
	}

	return revisions, nil
}
