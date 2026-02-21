package router

import (
	"github.com/go-chi/chi/v5"

	templatectl "github.com/codec404/chat-service/controller/template"
)

func registerTemplateRoutes(r chi.Router, h *templatectl.Handler) {
	r.Route("/templates", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{templateID}", h.Get)
		r.Put("/{templateID}", h.Update)
		r.Delete("/{templateID}", h.Delete)
	})
}
