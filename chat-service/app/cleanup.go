package app

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/codec404/chat-service/pkg/conc"
	log "github.com/codec404/chat-service/pkg/logger"
	tokenrepo "github.com/codec404/chat-service/repository/token"
)

// StartTokenCleanup launches a background goroutine that periodically deletes
// expired and revoked refresh tokens. It stops when ctx is cancelled (graceful shutdown).
func StartTokenCleanup(ctx context.Context, db *pgxpool.Pool, intervalHours int) {
	if intervalHours <= 0 {
		intervalHours = 6 // safe default
	}
	ticker := time.NewTicker(time.Duration(intervalHours) * time.Hour)
	repo := tokenrepo.NewPostgres(db)

	conc.SafeGo(TokenCleanupGoroutineName, func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				conc.SafeTry(TokenCleanupGoroutineName, func() {
					tickCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
					defer cancel()
					n, err := repo.DeleteExpiredAndRevoked(tickCtx)
					if err != nil {
						log.ErrorfWithContext(ctx, "token cleanup: %v", err)
					} else {
						log.InfofWithContext(ctx, "token cleanup: deleted %d expired/revoked tokens", n)
					}
				})
			case <-ctx.Done():
				return
			}
		}
	})
}
