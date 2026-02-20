package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	externalerror "github.com/codec404/chat-service/pkg/external_error"
	"github.com/codec404/chat-service/pkg/errorhandler"
	log "github.com/codec404/chat-service/pkg/logger"
	"github.com/codec404/chat-service/router"
)

func NewServer(cfg *Config) *http.Server {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(jsonRecoverer)

	router.GetAllRoutes(r)

	return &http.Server{
		Addr:              fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:           r,
		ReadTimeout:       ReadTimeout * time.Second,
		WriteTimeout:      WriteTimeout * time.Second,
		IdleTimeout:       IdleTimeout * time.Second,
		ReadHeaderTimeout: ReadHeaderTimeout * time.Second,
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
