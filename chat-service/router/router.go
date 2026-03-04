package router

import (
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/codec404/chat-service/controller/admin"
	authctl "github.com/codec404/chat-service/controller/auth"
	"github.com/codec404/chat-service/controller/campaign"
	"github.com/codec404/chat-service/controller/health"
	templatectl "github.com/codec404/chat-service/controller/template"
	"github.com/codec404/chat-service/controller/user"
	"github.com/codec404/chat-service/middleware"
	oauthrepo "github.com/codec404/chat-service/repository/oauth"
	tokenrepo "github.com/codec404/chat-service/repository/token"
	userrepo "github.com/codec404/chat-service/repository/user"
	authsvc "github.com/codec404/chat-service/service/auth"
)

func GetAllRoutes(r chi.Router, cfg Config) {
	uRepo := userrepo.NewPostgres(cfg.DB)
	tRepo := tokenrepo.NewPostgres(cfg.DB)
	oRepo := oauthrepo.NewPostgres(cfg.DB)

	googleOAuth2Config := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	aSvc := authsvc.New(uRepo, tRepo, oRepo, googleOAuth2Config, cfg.JWTSecret, cfg.AccessExpiryMin, cfg.RefreshExpiryDay)

	authenticate := middleware.Authenticate(cfg.JWTSecret)

	r.Route("/api/v1", func(r chi.Router) {
		// Public routes — no authentication required.
		registerHealthRoutes(r, health.NewHandler(cfg.DB))
		registerAuthRoutes(r, authctl.NewHandler(aSvc, googleOAuth2Config, cfg.SecureCookies, cfg.StateCookieMaxAge), cfg.AuthRPM)

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
