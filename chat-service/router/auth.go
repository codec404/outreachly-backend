package router

import (
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/codec404/chat-service/controller/auth"
	"github.com/codec404/chat-service/middleware"
)

func registerAuthRoutes(r chi.Router, h *auth.Handler, authRPM int) {
	rl := middleware.RateLimit(authRPM, time.Minute)

	r.Route("/auth", func(r chi.Router) {
		r.Use(rl)

		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
		r.Post("/refresh", h.Refresh)
		r.Post("/logout", h.Logout)
	})
}
