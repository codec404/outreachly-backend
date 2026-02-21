package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/codec404/chat-service/middleware"
	"github.com/codec404/chat-service/pkg/errorhandler"
	externalerror "github.com/codec404/chat-service/pkg/external_error"
	log "github.com/codec404/chat-service/pkg/logger"
	"github.com/codec404/chat-service/router"
)

func NewServer(cfg *Config, db *pgxpool.Pool) *http.Server {
	r := chi.NewRouter()

	// ── Core middleware (order matters) ──────────────────────────────────────
	r.Use(chimiddleware.RequestID)                             // generate request ID
	r.Use(setTraceID)                                          // bridge request ID → logger context
	r.Use(cors.Handler(buildCORSOptions(cfg.CORS)))            // CORS
	r.Use(bodyLimit(cfg.Server.MaxRequestBodyBytes))                     // body size guard
	r.Use(middleware.AccessLog)                                // structured access log
	r.Use(jsonRecoverer)                                       // panic → 500

	router.GetAllRoutes(r, router.Config{
		DB:               db,
		JWTSecret:        cfg.JWT.Secret,
		AccessExpiryMin:  cfg.JWT.AccessExpiry,
		RefreshExpiryDay: cfg.JWT.RefreshExpiry,
		AuthRPM:          cfg.RateLimit.AuthRPM,
	})

	return &http.Server{
		Addr:              fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:           r,
		ReadTimeout:       ReadTimeout * time.Second,
		WriteTimeout:      WriteTimeout * time.Second,
		IdleTimeout:       IdleTimeout * time.Second,
		ReadHeaderTimeout: ReadHeaderTimeout * time.Second,
	}
}

func buildCORSOptions(cfg CORSConfig) cors.Options {
	return cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"os-trace-id"},
		AllowCredentials: false,
		MaxAge:           300, // cache preflight for 5 minutes
	}
}

func StartServer(ctx context.Context, srv *http.Server) error {
	errCh := make(chan error, 1)

	go func() {
		log.Infof("server listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}

// setTraceID copies chi's RequestID into the logger context so every log call
// downstream automatically includes the trace_id field.
func setTraceID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := chimiddleware.GetReqID(r.Context())
		ctx := log.WithTraceID(r.Context(), traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func bodyLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

func jsonRecoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				log.ErrorfWithContext(r.Context(), "panic recovered: %v", rvr)
				errorhandler.Respond(w, r, externalerror.Internal())
			}
		}()
		next.ServeHTTP(w, r)
	})
}
