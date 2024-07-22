package handler_test

import (
	"context"

	"github.com/mirai-box/mirai-box/internal/middleware"
	"github.com/mirai-box/mirai-box/internal/model"
)

// Helper function to create a context with an authenticated user
func newContextWithUser(ctx context.Context, user *model.User) context.Context {
	return context.WithValue(ctx, middleware.UserKey, user)
}
