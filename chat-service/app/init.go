package app

import (
	"github.com/jackc/pgx/v5/pgxpool"

	log "github.com/codec404/chat-service/pkg/logger"
)

func Init() (*Config, *pgxpool.Pool, error) {
	log.InitLogger()

	if !isProduction() {
		loadEnvFile(LocalEnvFile) // soft fail — not present in prod
	}

	cfg, err := Load()
	if err != nil {
		return nil, nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, nil, err
	}

	if err := RunMigrations(cfg); err != nil {
		return nil, nil, err
	}

	if err := SeedSuperAdmin(cfg); err != nil {
		return nil, nil, err
	}

	db, err := OpenDB(cfg)
	if err != nil {
		return nil, nil, err
	}

	return cfg, db, nil
}
