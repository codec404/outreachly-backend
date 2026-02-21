package tokenrepo

import (
	"context"
	"time"

	"github.com/codec404/chat-service/model"
)

type Repository interface {
	Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) (*model.RefreshToken, error)
	FindByHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
	Revoke(ctx context.Context, id string) error
	RevokeAllForUser(ctx context.Context, userID string) error
}
