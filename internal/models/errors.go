package models

import "fmt"

// ErrNotFound represents a not found error
type ErrNotFound struct {
	Resource string
	ID       string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("%s with ID %s not found", e.Resource, e.ID)
}

