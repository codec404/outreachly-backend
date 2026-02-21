package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/httprate"

	"github.com/codec404/chat-service/pkg/errorhandler"
	externalerror "github.com/codec404/chat-service/pkg/external_error"
)

// RateLimit returns a per-IP rate limiter middleware.
// limit is the maximum number of requests allowed within the given window.
func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	return httprate.Limit(
		limit,
		window,
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			errorhandler.Respond(w, r, externalerror.New(http.StatusTooManyRequests, "too many requests, please try again later"))
		}),
	)
}
