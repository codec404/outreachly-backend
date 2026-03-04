package oauthrepo

import (
	"context"

	"github.com/codec404/chat-service/model"
)

type Repository interface {
	// FindByProvider returns an existing OAuth link by provider name + provider-issued user ID.
	// Returns externalerror.NotFound if no row exists.
	FindByProvider(ctx context.Context, provider, providerUserID string) (*model.OAuthProvider, error)

	// Create inserts a new row linking a local user to a provider identity.
	Create(ctx context.Context, p CreateParams) (*model.OAuthProvider, error)
}

type CreateParams struct {
	UserID         string
	Provider       string
	ProviderUserID string
	Email          string
	Name           string
	AvatarURL      string
}
