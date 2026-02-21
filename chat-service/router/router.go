package router

import (
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/codec404/chat-service/controller/admin"
	authctl "github.com/codec404/chat-service/controller/auth"
	"github.com/codec404/chat-service/controller/campaign"
	"github.com/codec404/chat-service/controller/health"
	templatectl "github.com/codec404/chat-service/controller/template"
	"github.com/codec404/chat-service/controller/user"
	"github.com/codec404/chat-service/middleware"
	tokenrepo "github.com/codec404/chat-service/repository/token"
	userrepo "github.com/codec404/chat-service/repository/user"
	authsvc "github.com/codec404/chat-service/service/auth"
)

func GetAllRoutes(r chi.Router, cfg Config) {
	uRepo := userrepo.NewPostgres(cfg.DB)
	tRepo := tokenrepo.NewPostgres(cfg.DB)
	aSvc := authsvc.New(uRepo, tRepo, cfg.JWTSecret, cfg.AccessExpiryMin, cfg.RefreshExpiryDay)

	authenticate := middleware.Authenticate(cfg.JWTSecret)

	r.Route("/api/v1", func(r chi.Router) {
		// Public routes — no authentication required.
		registerHealthRoutes(r, health.NewHandler(cfg.DB))
		registerAuthRoutes(r, authctl.NewHandler(aSvc), cfg.AuthRPM)

		// Authenticated routes — JWT required for everything below.
		r.Group(func(r chi.Router) {
			r.Use(authenticate)
			r.Use(middleware.RateLimitByUserID(cfg.UserRPM, time.Minute))

			registerUserRoutes(r, user.NewHandler())
			registerTemplateRoutes(r, templatectl.NewHandler())
			registerCampaignRoutes(r, campaign.NewHandler())
			registerAdminRoutes(r, admin.NewHandler())
		})
	})
}
