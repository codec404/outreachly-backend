package errorhandler

import (
	"encoding/json"
	"errors"
	"net/http"

	chimiddleware "github.com/go-chi/chi/v5/middleware"

	externalerror "github.com/codec404/chat-service/pkg/external_error"
	log "github.com/codec404/chat-service/pkg/logger"
)

const traceIDHeader = "os-trace-id"

type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func Respond(w http.ResponseWriter, r *http.Request, err error) {
	var extErr *externalerror.ExternalError
	if errors.As(err, &extErr) {
		if extErr.Unwrap() != nil {
			log.ErrorfWithContext(r.Context(), "request error: %v", extErr.Unwrap())
		}
		write(w, r, extErr.HTTPCode, extErr.Message)
		return
	}

	log.ErrorfWithContext(r.Context(), "unexpected error: %v", err)
	write(w, r, http.StatusInternalServerError, "internal server error")
}

func write(w http.ResponseWriter, r *http.Request, code int, message string) {
	traceID := log.TraceIDFromContext(r.Context())
	if traceID == "" {
		traceID = chimiddleware.GetReqID(r.Context())
	}

	w.Header().Set("Content-Type", "application/json")
	if traceID != "" {
		w.Header().Set(traceIDHeader, traceID)
	}
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(errorResponse{Code: code, Message: message})
}
