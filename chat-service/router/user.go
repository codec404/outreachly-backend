package router

import (
	"github.com/go-chi/chi/v5"

	"github.com/codec404/chat-service/controller/user"
)

func registerUserRoutes(r chi.Router, h *user.Handler) {
	r.Route("/users", func(r chi.Router) {
		r.Get("/me", h.GetMe)
		r.Put("/me", h.UpdateMe)
	})
}
