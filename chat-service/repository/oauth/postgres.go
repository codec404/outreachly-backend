package oauthrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/codec404/chat-service/model"
	externalerror "github.com/codec404/chat-service/pkg/external_error"
)

type Postgres struct {
	db *pgxpool.Pool
}

func NewPostgres(db *pgxpool.Pool) *Postgres {
	return &Postgres{db: db}
}

func (p *Postgres) FindByProvider(ctx context.Context, provider, providerUserID string) (*model.OAuthProvider, error) {
	o := &model.OAuthProvider{}
	err := p.db.QueryRow(ctx, `
		SELECT id, user_id, provider, provider_user_id, email, name, avatar_url, created_at
		FROM oauth_providers
		WHERE provider = $1 AND provider_user_id = $2
	`, provider, providerUserID).Scan(
		&o.ID, &o.UserID, &o.Provider, &o.ProviderUserID,
		&o.Email, &o.Name, &o.AvatarURL, &o.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, externalerror.NotFound("oauth provider link not found")
		}
		return nil, fmt.Errorf("oauthrepo.FindByProvider: %w", err)
	}
	return o, nil
}

func (p *Postgres) Create(ctx context.Context, params CreateParams) (*model.OAuthProvider, error) {
	o := &model.OAuthProvider{}
	err := p.db.QueryRow(ctx, `
		INSERT INTO oauth_providers (user_id, provider, provider_user_id, email, name, avatar_url)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, provider, provider_user_id, email, name, avatar_url, created_at
	`, params.UserID, params.Provider, params.ProviderUserID,
		params.Email, params.Name, params.AvatarURL,
	).Scan(
		&o.ID, &o.UserID, &o.Provider, &o.ProviderUserID,
		&o.Email, &o.Name, &o.AvatarURL, &o.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("oauthrepo.Create: %w", err)
	}
	return o, nil
}
