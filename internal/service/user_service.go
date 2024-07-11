package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/mirai-box/mirai-box/internal/models"
	"github.com/mirai-box/mirai-box/internal/repos"
)

// UserService implements the UserServiceInterface
type UserService struct {
	repo repos.UserRepositoryInterface
}

// NewUserService creates a new UserService
func NewUserService(repo repos.UserRepositoryInterface) UserServiceInterface {
	return &UserService{repo: repo}
}

// Authenticate verifies user credentials and returns the user if valid
func (s *UserService) Authenticate(ctx context.Context, username, password string) (*models.User, error) {
	slog.InfoContext(ctx, "Authenticating user", "username", username)

	user, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find user", "error", err, "username", username)
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		slog.ErrorContext(ctx, "Invalid credentials", "error", err, "username", username)
		return nil, errors.New("invalid credentials")
	}

	slog.InfoContext(ctx, "User authenticated successfully", "username", username)
	return user, nil
}

// FindByID retrieves a user by their ID
func (s *UserService) FindByID(ctx context.Context, id string) (*models.User, error) {
	slog.InfoContext(ctx, "Finding user by ID", "userID", id)

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find user by ID", "error", err, "userID", id)
		return nil, err
	}

	slog.InfoContext(ctx, "User found successfully", "userID", id)
	return user, nil
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, username, password, role string) (*models.User, error) {
	slog.InfoContext(ctx, "Creating new user", "username", username, "role", role)

	hashedPassword, err := HashPassword(password)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to hash password", "error", err)
		return nil, err
	}

	user := &models.User{
		ID:       uuid.New(),
		Username: username,
		Password: hashedPassword,
		Role:     role,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		slog.ErrorContext(ctx, "Failed to create user", "error", err, "username", username)
		return nil, err
	}

	slog.InfoContext(ctx, "User created successfully", "userID", user.ID, "username", username)
	return user, nil
}

// HashPassword hashes the password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (s *UserService) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	existingUser, err := s.repo.FindByID(ctx, user.ID.String())
	if err != nil {
		return nil, err
	}

	// Update only allowed fields
	existingUser.Username = user.Username
	existingUser.Role = user.Role

	// If a new password is provided, hash it
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		existingUser.Password = string(hashedPassword)
	}

	if err := s.repo.Update(ctx, existingUser); err != nil {
		return nil, err
	}

	return existingUser, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// Ensure UserService implements UserServiceInterface
var _ UserServiceInterface = (*UserService)(nil)
