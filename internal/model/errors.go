package model

import (
	"errors"
	"fmt"
)

// ErrNotFound represents a not found error
type ErrNotFound struct {
	Resource string
	ID       string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("%s with ID %s not found", e.Resource, e.ID)
}

var (
	ErrWebPageNotFound     = errors.New("webpage not found")
	ErrArtLinkNotFound     = errors.New("artlinl not found")
	ErrArtProjectNotFound  = errors.New("art project not found")
	ErrCollectionNotFound  = errors.New("collection not found")
	ErrStashNotFound       = errors.New("stash not found")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidInput        = errors.New("invalid input parameters")
	ErrStashCreationFailed = errors.New("failed to create stash")
	ErrUnauthorized        = errors.New("unauthorized access")
	ErrDuplicateUsername   = errors.New("username already exists")
	ErrRevisionNotFound    = errors.New("revision not found")
)
