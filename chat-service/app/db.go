package app

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// OpenDB creates a pgxpool connection with the settings from config.yml and
// verifies connectivity with a ping.
func OpenDB(cfg *Config) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=%s",
		SQLDriver, cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name, cfg.DB.SSLMode)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.ParseConfig: %w", err)
	}

	p := cfg.DB.Pool
	poolCfg.MaxConns = p.MaxConns
	poolCfg.MinConns = p.MinConns
	poolCfg.MaxConnIdleTime = time.Duration(p.MaxConnIdleMinutes) * time.Minute
	poolCfg.MaxConnLifetime = time.Duration(p.MaxConnLifetimeHours) * time.Hour
	poolCfg.HealthCheckPeriod = 30 * time.Second
	poolCfg.ConnConfig.ConnectTimeout = 5 * time.Second

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.NewWithConfig: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db ping: %w", err)
	}

	return pool, nil
}
