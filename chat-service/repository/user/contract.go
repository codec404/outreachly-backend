package userrepo

import (
	"context"

	"github.com/codec404/chat-service/model"
)

type Repository interface {
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
	CreateWithRole(ctx context.Context, name, email, passwordHash, roleName string) (*model.User, error)
	GetRoles(ctx context.Context, userID string) ([]string, error)
}
