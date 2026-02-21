package app

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"

	"github.com/codec404/chat-service/db/migrations"
	log "github.com/codec404/chat-service/pkg/logger"
)

func RunMigrations(cfg *Config) error {
	db, err := sql.Open(SQLDriver, buildDSN(cfg.DB))
	if err != nil {
		return fmt.Errorf("open db for migrations: %w", err)
	}
	defer db.Close()

	driver, err := migratepostgres.WithInstance(db, &migratepostgres.Config{})
	if err != nil {
		return fmt.Errorf("migration driver: %w", err)
	}

	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("migration source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, cfg.DB.Name, driver)
	if err != nil {
		return fmt.Errorf("migration instance: %w", err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Infof("migrations: already up to date")
			return nil
		}
		return fmt.Errorf("migrate up: %w", err)
	}

	log.Infof("migrations: applied successfully")
	return nil
}

func buildDSN(db DBConfig) string {
	return fmt.Sprintf(
		"%s://%s:%s@%s:%s/%s?sslmode=%s",
		SQLDriver, db.User, db.Password, db.Host, db.Port, db.Name, db.SSLMode,
	)
}
