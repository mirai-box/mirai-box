package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/mirai-box/mirai-box/internal/models"
	"github.com/mirai-box/mirai-box/internal/repos"
)

// SaleService implements the SaleServiceInterface
type saleService struct {
	repo repos.SaleRepositoryInterface
}

// NewSaleService creates a new instance of SaleService
func NewSaleService(repo repos.SaleRepositoryInterface) SaleServiceInterface {
	return &saleService{repo: repo}
}

// CreateSale creates a new sale record for an art project
func (s *saleService) CreateSale(ctx context.Context, artProjectID, userID string, price float64) (*models.Sale, error) {
	sale := &models.Sale{
		ID:           uuid.New(),
		ArtProjectID: uuid.MustParse(artProjectID),
		UserID:       uuid.MustParse(userID),
		Price:        price,
		SoldAt:       time.Now(),
	}

	slog.InfoContext(ctx, "Creating new sale", "artProjectID", artProjectID, "userID", userID, "price", price)

	if err := s.repo.Create(ctx, sale); err != nil {
		slog.ErrorContext(ctx, "Failed to create sale", "error", err, "artProjectID", artProjectID, "userID", userID)
		return nil, err
	}

	slog.InfoContext(ctx, "Sale created successfully", "saleID", sale.ID, "artProjectID", artProjectID)
	return sale, nil
}

// FindByID finds a sale by its ID
func (s *saleService) FindByID(ctx context.Context, id string) (*models.Sale, error) {
	slog.InfoContext(ctx, "Finding sale by ID", "saleID", id)

	sale, err := s.repo.FindByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find sale by ID", "error", err, "saleID", id)
		return nil, err
	}

	slog.InfoContext(ctx, "Sale found successfully", "saleID", id)
	return sale, nil
}

// FindByUserID finds all sales made by a specific user
func (s *saleService) FindByUserID(ctx context.Context, userID string) ([]models.Sale, error) {
	slog.InfoContext(ctx, "Finding sales by user ID", "userID", userID)

	sales, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find sales by user ID", "error", err, "userID", userID)
		return nil, err
	}

	slog.InfoContext(ctx, "Sales found successfully", "userID", userID, "count", len(sales))
	return sales, nil
}

// FindByArtProjectID finds all sales for a specific art project
func (s *saleService) FindByArtProjectID(ctx context.Context, artProjectID string) ([]models.Sale, error) {
	slog.InfoContext(ctx, "Finding sales by art project ID", "artProjectID", artProjectID)

	sales, err := s.repo.FindByArtProjectID(ctx, artProjectID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find sales by art project ID", "error", err, "artProjectID", artProjectID)
		return nil, err
	}

	slog.InfoContext(ctx, "Sales found successfully", "artProjectID", artProjectID, "count", len(sales))
	return sales, nil
}

// Ensure SaleService implements SaleServiceInterface
var _ SaleServiceInterface = (*saleService)(nil)
