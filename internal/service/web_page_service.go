package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/mirai-box/mirai-box/internal/models"
	"github.com/mirai-box/mirai-box/internal/repos"
)

// WebPageService implements the WebPageServiceInterface
type WebPageService struct {
	repo repos.WebPageRepositoryInterface
}

// NewWebPageService creates a new WebPageService
func NewWebPageService(repo repos.WebPageRepositoryInterface) WebPageServiceInterface {
	return &WebPageService{repo: repo}
}

// CreateWebPage creates a new web page
func (s *WebPageService) CreateWebPage(ctx context.Context, webPage *models.WebPage) (*models.WebPage, error) {
	slog.InfoContext(ctx, "Creating new web page",
		"userID", webPage.UserID,
		"title", webPage.Title,
		"pageType", webPage.PageType,
	)

	wp := &models.WebPage{
		ID:        uuid.New(),
		UserID:    webPage.UserID,
		Title:     webPage.Title,
		Html:      webPage.Html,
		PageType:  webPage.PageType,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, wp); err != nil {
		slog.ErrorContext(ctx, "Failed to create web page", "error", err)
		return nil, err
	}

	slog.InfoContext(ctx, "Web page created successfully", "pageID", wp.ID)
	return wp, nil
}

// UpdateWebPage updates an existing web page
func (s *WebPageService) UpdateWebPage(ctx context.Context, webPage *models.WebPage) (*models.WebPage, error) {
	slog.InfoContext(ctx, "Updating web page", "pageID", webPage.ID)

	webPage.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, webPage); err != nil {
		slog.ErrorContext(ctx, "Failed to update web page", "error", err, "pageID", webPage.ID)
		return nil, err
	}

	slog.InfoContext(ctx, "Web page updated successfully", "pageID", webPage.ID)
	return webPage, nil
}

// DeleteWebPage deletes a web page by ID
func (s *WebPageService) DeleteWebPage(ctx context.Context, id string) error {
	slog.InfoContext(ctx, "Deleting web page", "pageID", id)

	if err := s.repo.Delete(ctx, uuid.MustParse(id)); err != nil {
		slog.ErrorContext(ctx, "Failed to delete web page", "error", err, "pageID", id)
		return err
	}

	slog.InfoContext(ctx, "Web page deleted successfully", "pageID", id)
	return nil
}

// GetWebPage retrieves a web page by ID
func (s *WebPageService) GetWebPage(ctx context.Context, id string) (*models.WebPage, error) {
	slog.InfoContext(ctx, "Getting web page", "pageID", id)

	wp, err := s.repo.FindByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get web page", "error", err, "pageID", id)
		return nil, err
	}

	slog.InfoContext(ctx, "Web page retrieved successfully", "pageID", id)
	return wp, nil
}

// ListWebPages retrieves all web pages
func (s *WebPageService) ListWebPages(ctx context.Context) ([]models.WebPage, error) {
	slog.InfoContext(ctx, "Listing all web pages")

	pages, err := s.repo.FindAll(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to list web pages", "error", err)
		return nil, err
	}

	slog.InfoContext(ctx, "Web pages listed successfully", "count", len(pages))
	return pages, nil
}

// ListWebPagesByType retrieves all web pages by type
func (s *WebPageService) ListWebPagesByType(ctx context.Context, pageType string) ([]models.WebPage, error) {
	slog.InfoContext(ctx, "Listing web pages by type", "pageType", pageType)

	pages, err := s.repo.FindByType(ctx, pageType)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to list web pages by type", "error", err, "pageType", pageType)
		return nil, err
	}

	slog.InfoContext(ctx, "Web pages listed successfully", "pageType", pageType, "count", len(pages))
	return pages, nil
}

func (s *WebPageService) ListUserWebPages(ctx context.Context, userID string) ([]models.WebPage, error) {
	slog.InfoContext(ctx, "Listing web pages by user", "userID", userID)

	pages, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to list web pages by user", "error", err, "userID", userID)
		return nil, err
	}

	slog.InfoContext(ctx, "Web pages listed successfully", "userID", userID, "count", len(pages))
	return pages, nil
}

// Ensure WebPageService implements WebPageServiceInterface
var _ WebPageServiceInterface = (*WebPageService)(nil)
