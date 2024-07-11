package repos

import (
	"context"
	"log/slog"

	"gorm.io/gorm"

	"github.com/mirai-box/mirai-box/internal/models"
)

// UserRepository implements the UserRepositoryInterface
type userRepository struct {
	DB *gorm.DB
}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository(db *gorm.DB) UserRepositoryInterface {
	return &userRepository{DB: db}
}

// Create adds a new user to the database
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	slog.InfoContext(ctx, "Creating new user", "userID", user.ID, "username", user.Username)
	if err := r.DB.Create(user).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to create user", "error", err, "userID", user.ID, "username", user.Username)
		return err
	}
	slog.InfoContext(ctx, "User created successfully", "userID", user.ID, "username", user.Username)
	return nil
}

// FindByID retrieves a user by their ID
func (r *userRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	slog.InfoContext(ctx, "Finding user by ID", "userID", id)
	var user models.User
	if err := r.DB.First(&user, "id = ?", id).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to find user by ID", "error", err, "userID", id)
		return nil, err
	}
	slog.InfoContext(ctx, "User found successfully", "userID", id, "username", user.Username)
	return &user, nil
}

// FindByUsername retrieves a user by their username
func (r *userRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	slog.InfoContext(ctx, "Finding user by username", "username", username)
	var user models.User
	if err := r.DB.First(&user, "username = ?", username).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to find user by username", "error", err, "username", username)
		return nil, err
	}
	slog.InfoContext(ctx, "User found successfully", "userID", user.ID, "username", username)
	return &user, nil
}

// Update modifies an existing user in the database
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	slog.InfoContext(ctx, "Updating user", "userID", user.ID, "username", user.Username)
	if err := r.DB.Save(user).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to update user", "error", err, "userID", user.ID, "username", user.Username)
		return err
	}
	slog.InfoContext(ctx, "User updated successfully", "userID", user.ID, "username", user.Username)
	return nil
}

// Delete removes a user from the database
func (r *userRepository) Delete(ctx context.Context, id string) error {
	slog.InfoContext(ctx, "Deleting user", "userID", id)
	if err := r.DB.Delete(&models.User{}, "id = ?", id).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to delete user", "error", err, "userID", id)
		return err
	}
	slog.InfoContext(ctx, "User deleted successfully", "userID", id)
	return nil
}

// Ensure UserRepository implements UserRepositoryInterface
var _ UserRepositoryInterface = (*userRepository)(nil)
