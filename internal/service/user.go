package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/mirai-box/mirai-box/internal/model"
	"github.com/mirai-box/mirai-box/internal/repo"
)

//go:generate go run github.com/vektra/mockery/v2@v2 --name=UserService --filename=user_service.go --output=../../mocks/
type UserService interface {
	Authenticate(ctx context.Context, username, password string) (*model.User, error)
	CreateUser(ctx context.Context, username, password, role string) (*model.User, error)
	GetUser(ctx context.Context, id string) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, id string) error
	GetStorageUsage(ctx context.Context, userID string) (*model.StorageUsage, error)
	UpdateStorageUsage(ctx context.Context, storageUsage *model.StorageUsage) error
	GetStashByUserID(ctx context.Context, userID string) (*model.Stash, error)
}

// UserService implements the UserServiceInterface
type userService struct {
	userRepo repo.UserRepository
}

// NewUserService creates a new UserService
func NewUserService(repo repo.UserRepository) UserService {
	return &userService{userRepo: repo}
}

// Authenticate verifies user credentials and returns the user if valid
func (s *userService) Authenticate(ctx context.Context, username, password string) (*model.User, error) {
	slog.InfoContext(ctx, "Authenticating user", "username", username)

	user, err := s.userRepo.FindUserByUsername(ctx, username)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find user", "error", err, "username", username)
		return nil, err
	}
	if user == nil {
		slog.ErrorContext(ctx, "Failed to find user", "username", username)
		return nil, fmt.Errorf("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		slog.ErrorContext(ctx, "Invalid credentials", "error", err, "username", username)
		return nil, errors.New("invalid credentials")
	}

	slog.InfoContext(ctx, "User authenticated successfully", "username", username)
	return user, nil
}

// GetUser retrieves a user by their ID
func (s *userService) GetUser(ctx context.Context, id string) (*model.User, error) {
	slog.InfoContext(ctx, "Finding user by ID", "userID", id)

	user, err := s.userRepo.FindUserByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find user by ID", "error", err, "userID", id)
		return nil, err
	}

	slog.InfoContext(ctx, "User found successfully", "userID", id)
	return user, nil
}

func (s *userService) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	return s.userRepo.FindUserByUsername(ctx, username)
}

// CreateUser creates a new user
func (s *userService) CreateUser(ctx context.Context, username, password, role string) (*model.User, error) {
	slog.InfoContext(ctx, "Creating new user", "username", username, "role", role)

	hashedPassword, err := HashPassword(password)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to hash password", "error", err)
		return nil, err
	}

	user := &model.User{
		ID:       uuid.New(),
		Username: username,
		Password: hashedPassword,
		Role:     role,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		slog.ErrorContext(ctx, "Failed to create user", "error", err, "username", username)
		return nil, err
	}

	if err := s.createStash(ctx, user.ID); err != nil {
		slog.ErrorContext(ctx, "Failed to create user", "error", err, "username", username)
		return nil, err
	}

	slog.InfoContext(ctx, "User created successfully", "userID", user.ID, "username", username)
	return user, nil
}

func (s *userService) createStash(ctx context.Context, userID uuid.UUID) error {
	slog.InfoContext(ctx, "Creating new stash", "userID", userID)

	stash := &model.Stash{
		ID:     uuid.New(),
		UserID: userID,
	}

	if err := s.userRepo.CreateStash(ctx, stash); err != nil {
		slog.ErrorContext(ctx, "Failed to create stash", "error", err, "userID", userID.String())
		return err
	}

	slog.InfoContext(ctx, "Stash created successfully", "stashID", stash.ID, "userID", userID)
	return nil
}

func (s *userService) UpdateUser(ctx context.Context, user *model.User) error {
	existingUser, err := s.userRepo.FindUserByID(ctx, user.ID.String())
	if err != nil {
		return err
	}

	// Update only allowed fields
	existingUser.Username = user.Username
	existingUser.Role = user.Role

	// If a new password is provided, hash it
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		existingUser.Password = string(hashedPassword)
	}

	if err := s.userRepo.UpdateUser(ctx, existingUser); err != nil {
		return err
	}

	return nil
}

func (s *userService) DeleteUser(ctx context.Context, id string) error {
	return s.userRepo.DeleteUser(ctx, id)
}

func (s *userService) GetStorageUsage(ctx context.Context, userID string) (*model.StorageUsage, error) {
	return s.userRepo.GetStorageUsage(ctx, userID)
}

func (s *userService) UpdateStorageUsage(ctx context.Context, storageUsage *model.StorageUsage) error {
	return s.userRepo.UpdateStorageUsage(ctx, storageUsage)
}

func (s *userService) GetStashByUserID(ctx context.Context, userID string) (*model.Stash, error) {
	return s.userRepo.GetStashByUserID(ctx, userID)
}
