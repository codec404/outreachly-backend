package router

import (
	"github.com/go-chi/chi/v5"

	"github.com/codec404/chat-service/controller/admin"
	"github.com/codec404/chat-service/middleware"
)

func registerAdminRoutes(r chi.Router, h *admin.Handler) {
	r.Route("/admin", func(r chi.Router) {
		r.Use(middleware.RequireRole("admin", "super_admin"))

		// User management (admin + super_admin)
		r.Get("/users", h.ListUsers)
		r.Get("/users/{userID}", h.GetUser)
		r.Put("/users/{userID}/block", h.BlockUser)
		r.Put("/users/{userID}/unblock", h.UnblockUser)
		r.Delete("/users/{userID}", h.DeleteUser)

		// Role assignment (super_admin only)
		r.With(middleware.RequireRole("super_admin")).Put("/users/{userID}/role", h.UpdateUserRole)
	})
}
