package router

import (
	"github.com/go-chi/chi/v5"

	"github.com/codec404/chat-service/controller/health"
)

func registerHealthRoutes(r chi.Router, h *health.Handler) {
	r.Get("/health/live", h.Live)
	r.Get("/health/ready", h.Ready)
}
