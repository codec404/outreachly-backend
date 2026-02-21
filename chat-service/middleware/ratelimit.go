package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/httprate"

	"github.com/codec404/chat-service/pkg/errorhandler"
	externalerror "github.com/codec404/chat-service/pkg/external_error"
)

// RateLimit returns a per-IP rate limiter middleware.
func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	return httprate.Limit(
		limit,
		window,
		httprate.WithLimitHandler(tooManyRequests),
	)
}

// RateLimitByUserID returns a per-user rate limiter for authenticated routes.
// Keys by user ID from context; falls back to IP if no user is present.
func RateLimitByUserID(limit int, window time.Duration) func(http.Handler) http.Handler {
	return httprate.Limit(
		limit,
		window,
		httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
			if userID, ok := UserIDFromContext(r.Context()); ok {
				return "uid:" + userID, nil
			}
			return httprate.KeyByIP(r)
		}),
		httprate.WithLimitHandler(tooManyRequests),
	)
}

func tooManyRequests(w http.ResponseWriter, r *http.Request) {
	errorhandler.Respond(w, r, externalerror.New(http.StatusTooManyRequests, "too many requests, please try again later"))
}
