package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func GetAllRoutes(r chi.Router) {
	r.Get("/ping", pingHandler)
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}
