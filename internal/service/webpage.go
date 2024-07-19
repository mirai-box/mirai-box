package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/mirai-box/mirai-box/internal/model"
	"github.com/mirai-box/mirai-box/internal/repo"
)

//go:generate go run github.com/vektra/mockery/v2@v2 --name=WebPageService --filename=webpage_service.go --output=../../mocks/
type WebPageService interface {
	CreateWebPage(ctx context.Context, webPage *model.WebPage) (*model.WebPage, error)
	UpdateWebPage(ctx context.Context, webPage *model.WebPage) (*model.WebPage, error)
	DeleteWebPage(ctx context.Context, id string) error
	GetWebPage(ctx context.Context, id string) (*model.WebPage, error)
	ListWebPages(ctx context.Context) ([]model.WebPage, error)
	ListUserWebPages(ctx context.Context, userID string) ([]model.WebPage, error)
	ListWebPagesByType(ctx context.Context, pageType string) ([]model.WebPage, error)
}

// WebPageService implements the WebPageServiceInterface
type webPageService struct {
	repo repo.WebPageRepository
}

// NewWebPageService creates a new WebPageService
func NewWebPageService(repo repo.WebPageRepository) WebPageService {
	return &webPageService{repo: repo}
}

// CreateWebPage creates a new web page
func (s *webPageService) CreateWebPage(ctx context.Context, webPage *model.WebPage) (*model.WebPage, error) {
	slog.InfoContext(ctx, "Creating new web page",
		"userID", webPage.UserID,
		"title", webPage.Title,
		"pageType", webPage.PageType,
	)

	wp := &model.WebPage{
		ID:        uuid.New(),
		UserID:    webPage.UserID,
		Title:     webPage.Title,
		Html:      webPage.Html,
		PageType:  webPage.PageType,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateWebPage(ctx, wp); err != nil {
		slog.ErrorContext(ctx, "Failed to create web page", "error", err)
		return nil, err
	}

	slog.InfoContext(ctx, "Web page created successfully", "pageID", wp.ID)
	return wp, nil
}

// UpdateWebPage updates an existing web page
func (s *webPageService) UpdateWebPage(ctx context.Context, webPage *model.WebPage) (*model.WebPage, error) {
	slog.InfoContext(ctx, "Updating web page", "pageID", webPage.ID)

	webPage.UpdatedAt = time.Now()

	if err := s.repo.UpdateWebPage(ctx, webPage); err != nil {
		slog.ErrorContext(ctx, "Failed to update web page", "error", err, "pageID", webPage.ID)
		return nil, err
	}

	slog.InfoContext(ctx, "Web page updated successfully", "pageID", webPage.ID)
	return webPage, nil
}

// DeleteWebPage deletes a web page by ID
func (s *webPageService) DeleteWebPage(ctx context.Context, id string) error {
	slog.InfoContext(ctx, "Deleting web page", "pageID", id)

	if err := s.repo.DeleteWebPage(ctx, id); err != nil {
		slog.ErrorContext(ctx, "Failed to delete web page", "error", err, "pageID", id)
		return err
	}

	slog.InfoContext(ctx, "Web page deleted successfully", "pageID", id)
	return nil
}

// GetWebPage retrieves a web page by ID
func (s *webPageService) GetWebPage(ctx context.Context, id string) (*model.WebPage, error) {
	slog.InfoContext(ctx, "Getting web page", "pageID", id)

	wp, err := s.repo.FindWebPageByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get web page", "error", err, "pageID", id)
		return nil, err
	}

	slog.InfoContext(ctx, "Web page retrieved successfully", "pageID", id)
	return wp, nil
}

// ListWebPages retrieves all web pages
func (s *webPageService) ListWebPages(ctx context.Context) ([]model.WebPage, error) {
	slog.InfoContext(ctx, "Listing all web pages")

	pages, err := s.repo.FindAllWebPages(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to list web pages", "error", err)
		return nil, err
	}

	slog.InfoContext(ctx, "Web pages listed successfully", "count", len(pages))
	return pages, nil
}

// ListWebPagesByType retrieves all web pages by type
func (s *webPageService) ListWebPagesByType(ctx context.Context, pageType string) ([]model.WebPage, error) {
	slog.InfoContext(ctx, "Listing web pages by type", "pageType", pageType)

	pages, err := s.repo.FindWebPagesByType(ctx, pageType)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to list web pages by type", "error", err, "pageType", pageType)
		return nil, err
	}

	slog.InfoContext(ctx, "Web pages listed successfully", "pageType", pageType, "count", len(pages))
	return pages, nil
}

func (s *webPageService) ListUserWebPages(ctx context.Context, userID string) ([]model.WebPage, error) {
	slog.InfoContext(ctx, "Listing web pages by user", "userID", userID)

	pages, err := s.repo.FindWebPagesByUserID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to list web pages by user", "error", err, "userID", userID)
		return nil, err
	}

	slog.InfoContext(ctx, "Web pages listed successfully", "userID", userID, "count", len(pages))
	return pages, nil
}
