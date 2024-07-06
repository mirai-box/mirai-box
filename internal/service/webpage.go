package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mirai-box/mirai-box/internal/model"
	"github.com/mirai-box/mirai-box/internal/repository"
)

// webPageService implements the WebPageService interface.
type webPageService struct {
	repo repository.WebPageRepository
}

// NewWebPageService creates a new WebPageService.
func NewWebPageService(repo repository.WebPageRepository) WebPageService {
	return &webPageService{repo: repo}
}

// CreateWebPage creates a new web page.
func (s *webPageService) CreateWebPage(ctx context.Context, title, html string) (*model.WebPage, error) {
	wp := &model.WebPage{
		ID:        uuid.New().String(),
		Title:     title,
		HTML:      html,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.repo.CreateWebPage(context.Background(), wp)
	if err != nil {
		return nil, err
	}

	return wp, nil
}

// UpdateWebPage updates an existing web page.
func (s *webPageService) UpdateWebPage(ctx context.Context, id, title, html string) (*model.WebPage, error) {
	wp := &model.WebPage{
		ID:        id,
		Title:     title,
		HTML:      html,
		UpdatedAt: time.Now(),
	}

	err := s.repo.UpdateWebPage(context.Background(), wp)
	if err != nil {
		return nil, err
	}

	return wp, nil
}

// DeleteWebPage deletes a web page by ID.
func (s *webPageService) DeleteWebPage(ctx context.Context, id string) error {
	return s.repo.DeleteWebPage(context.Background(), id)
}

// GetWebPage retrieves a web page by ID.
func (s *webPageService) GetWebPage(ctx context.Context, id string) (*model.WebPage, error) {
	return s.repo.GetWebPage(context.Background(), id)
}

// ListWebPages retrieves all web pages.
func (s *webPageService) ListWebPages(ctx context.Context) ([]model.WebPage, error) {
	return s.repo.ListWebPages(context.Background())
}

// ListWebPages retrieves all web pages.
func (s *webPageService) ListWebPagesByType(ctx context.Context, webPagesType string) ([]model.WebPage, error) {
	return s.repo.ListWebPagesByType(ctx, webPagesType)
}
