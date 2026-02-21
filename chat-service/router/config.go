package router

import "github.com/jackc/pgx/v5/pgxpool"

// Config holds the dependencies needed to wire up all route handlers.
type Config struct {
	DB               *pgxpool.Pool
	JWTSecret        string
	AccessExpiryMin  int
	RefreshExpiryDay int
}
