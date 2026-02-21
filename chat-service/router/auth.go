package router

import (
	"github.com/go-chi/chi/v5"

	"github.com/codec404/chat-service/controller/auth"
)

func registerAuthRoutes(r chi.Router, h *auth.Handler) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
		r.Post("/refresh", h.Refresh)
		r.Post("/logout", h.Logout)
	})
}
