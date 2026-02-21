package app

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	log "github.com/codec404/chat-service/pkg/logger"
)

func SeedSuperAdmin(cfg *Config) error {
	db, err := sql.Open(SQLDriver, buildDSN(cfg.DB))
	if err != nil {
		return fmt.Errorf("open db for seeding: %w", err)
	}
	defer db.Close()

	// Idempotent: skip if a super_admin user already exists.
	var count int
	if err := db.QueryRow(`
		SELECT COUNT(*)
		FROM   user_roles ur
		JOIN   roles r ON r.id = ur.role_id
		WHERE  r.name = 'super_admin'
	`).Scan(&count); err != nil {
		return fmt.Errorf("check super admin existence: %w", err)
	}
	if count > 0 {
		log.Infof("seed: super admin already exists, skipping")
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.SuperAdmin.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash super admin password: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin seed transaction: %w", err)
	}
	defer tx.Rollback() // no-op after Commit; intentionally unhandled (sql.ErrTxDone is expected)

	var userID string
	if err := tx.QueryRow(`
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id
	`, cfg.SuperAdmin.Name, cfg.SuperAdmin.Email, string(hash)).Scan(&userID); err != nil {
		return fmt.Errorf("insert super admin user: %w", err)
	}

	if _, err := tx.Exec(`
		INSERT INTO user_roles (user_id, role_id)
		SELECT $1, id FROM roles WHERE name = 'super_admin'
	`, userID); err != nil {
		return fmt.Errorf("assign super admin role: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit super admin seed: %w", err)
	}

	log.Infof("seed: super admin created (email: %s)", cfg.SuperAdmin.Email)
	return nil
}
