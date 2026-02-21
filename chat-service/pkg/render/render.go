package render

import (
	"encoding/json"
	"net/http"

	chimiddleware "github.com/go-chi/chi/v5/middleware"

	log "github.com/codec404/chat-service/pkg/logger"
)

const traceIDHeader = "os-trace-id"

// JSONResponse writes a JSON-encoded body with the given HTTP status code.
// The trace ID is sent as the "os-trace-id" response header.
func JSONResponse(w http.ResponseWriter, r *http.Request, status int, data any) {
	traceID := log.TraceIDFromContext(r.Context())
	if traceID == "" {
		traceID = chimiddleware.GetReqID(r.Context())
	}

	w.Header().Set("Content-Type", "application/json")
	if traceID != "" {
		w.Header().Set(traceIDHeader, traceID)
	}
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
