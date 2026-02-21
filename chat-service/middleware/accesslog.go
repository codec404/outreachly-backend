package middleware

import (
	"net/http"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"

	log "github.com/codec404/chat-service/pkg/logger"
)

// AccessLog logs each completed request as a structured log line via Zap.
// Includes method, path, status code, bytes written, duration, and trace ID.
func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		log.InfofWithContext(r.Context(),
			"%s %s %d %dB %s",
			r.Method,
			r.URL.RequestURI(),
			ww.Status(),
			ww.BytesWritten(),
			time.Since(start).Round(time.Millisecond),
		)
	})
}
