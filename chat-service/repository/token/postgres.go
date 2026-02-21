package tokenrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

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

func (p *Postgres) Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) (*model.RefreshToken, error) {
	t := &model.RefreshToken{}
	err := p.db.QueryRow(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, token_hash, expires_at, revoked_at, created_at
	`, userID, tokenHash, expiresAt).Scan(
		&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.RevokedAt, &t.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("tokenrepo.Create: %w", err)
	}
	return t, nil
}

func (p *Postgres) FindByHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	t := &model.RefreshToken{}
	err := p.db.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`, tokenHash).Scan(
		&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.RevokedAt, &t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, externalerror.Unauthorized("invalid refresh token")
		}
		return nil, fmt.Errorf("tokenrepo.FindByHash: %w", err)
	}
	return t, nil
}

// FindAndRevokeByHash atomically marks the token as revoked and returns it in one statement.
// It returns Unauthorized if the token doesn't exist, is already revoked, or is expired.
func (p *Postgres) FindAndRevokeByHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	t := &model.RefreshToken{}
	err := p.db.QueryRow(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE token_hash = $1
		  AND revoked_at IS NULL
		  AND expires_at > NOW()
		RETURNING id, user_id, token_hash, expires_at, revoked_at, created_at
	`, tokenHash).Scan(
		&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.RevokedAt, &t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, externalerror.Unauthorized("refresh token is invalid, expired, or already used")
		}
		return nil, fmt.Errorf("tokenrepo.FindAndRevokeByHash: %w", err)
	}
	return t, nil
}

func (p *Postgres) Revoke(ctx context.Context, id string) error {
	_, err := p.db.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = NOW() WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("tokenrepo.Revoke: %w", err)
	}
	return nil
}

func (p *Postgres) RevokeAllForUser(ctx context.Context, userID string) error {
	_, err := p.db.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = NOW()
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)
	if err != nil {
		return fmt.Errorf("tokenrepo.RevokeAllForUser: %w", err)
	}
	return nil
}

// DeleteExpiredAndRevoked removes tokens that have expired or been explicitly revoked.
// Called by the background cleanup goroutine.
func (p *Postgres) DeleteExpiredAndRevoked(ctx context.Context) (int64, error) {
	result, err := p.db.Exec(ctx, `
		DELETE FROM refresh_tokens
		WHERE expires_at < NOW() OR revoked_at IS NOT NULL
	`)
	if err != nil {
		return 0, fmt.Errorf("tokenrepo.DeleteExpiredAndRevoked: %w", err)
	}
	return result.RowsAffected(), nil
}
