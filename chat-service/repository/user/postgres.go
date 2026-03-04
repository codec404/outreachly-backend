package userrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

func (p *Postgres) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	u := &model.User{}
	err := p.db.QueryRow(ctx, `
		SELECT id, name, email, password_hash, avatar_url, is_active, is_blocked, created_at, updated_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`, email).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.AvatarURL, &u.IsActive, &u.IsBlocked, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, externalerror.NotFound("user not found")
		}
		return nil, fmt.Errorf("userrepo.FindByEmail: %w", err)
	}
	return u, nil
}

func (p *Postgres) FindByID(ctx context.Context, id string) (*model.User, error) {
	u := &model.User{}
	err := p.db.QueryRow(ctx, `
		SELECT id, name, email, password_hash, avatar_url, is_active, is_blocked, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.AvatarURL, &u.IsActive, &u.IsBlocked, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, externalerror.NotFound("user not found")
		}
		return nil, fmt.Errorf("userrepo.FindByID: %w", err)
	}
	return u, nil
}

func (p *Postgres) CreateWithRole(ctx context.Context, name, email, passwordHash, roleName string) (*model.User, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("userrepo.CreateWithRole begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	u := &model.User{}
	err = tx.QueryRow(ctx, `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, password_hash, avatar_url, is_active, is_blocked, created_at, updated_at
	`, name, email, passwordHash).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.AvatarURL, &u.IsActive, &u.IsBlocked, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if isDuplicateError(err) {
			return nil, externalerror.Conflict("email already in use")
		}
		return nil, fmt.Errorf("userrepo.CreateWithRole insert user: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id)
		SELECT $1, id FROM roles WHERE name = $2
	`, u.ID, roleName)
	if err != nil {
		return nil, fmt.Errorf("userrepo.CreateWithRole assign role: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("userrepo.CreateWithRole commit: %w", err)
	}
	return u, nil
}

func (p *Postgres) GetRoles(ctx context.Context, userID string) ([]string, error) {
	rows, err := p.db.Query(ctx, `
		SELECT r.name
		FROM roles r
		JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("userrepo.GetRoles: %w", err)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, fmt.Errorf("userrepo.GetRoles scan: %w", err)
		}
		roles = append(roles, role)
	}
	// rows.Err() is nil on an empty result set; pgx only sets it on iteration failure.
	return roles, rows.Err()
}

func (p *Postgres) UpdateAvatarURL(ctx context.Context, userID, avatarURL string) error {
	_, err := p.db.Exec(ctx, `
		UPDATE users SET avatar_url = $1 WHERE id = $2 AND deleted_at IS NULL
	`, avatarURL, userID)
	if err != nil {
		return fmt.Errorf("userrepo.UpdateAvatarURL: %w", err)
	}
	return nil
}

func isDuplicateError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
