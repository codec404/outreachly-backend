package errorhandler

import (
	"encoding/json"
	"errors"
	"net/http"

	externalerror "github.com/codec404/chat-service/pkg/external_error"
	log "github.com/codec404/chat-service/pkg/logger"
)

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
		write(w, extErr.HTTPCode, extErr.Message)
		return
	}

	log.ErrorfWithContext(r.Context(), "unexpected error: %v", err)
	write(w, http.StatusInternalServerError, "internal server error")
}

func write(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(errorResponse{Code: code, Message: message})
}
