package router

import (
	"github.com/go-chi/chi/v5"

	"github.com/codec404/chat-service/controller/campaign"
)

func registerCampaignRoutes(r chi.Router, h *campaign.Handler) {
	r.Route("/campaigns", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{campaignID}", h.Get)
		r.Put("/{campaignID}", h.Update)
		r.Delete("/{campaignID}", h.Delete)

		// Campaign lifecycle actions
		r.Post("/{campaignID}/start", h.Start)
		r.Post("/{campaignID}/schedule", h.Schedule)

		// Target management
		r.Get("/{campaignID}/targets", h.ListTargets)
		r.Post("/{campaignID}/targets", h.AddTargets)
	})
}
