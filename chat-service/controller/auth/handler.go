package auth

import (
	"encoding/json"
	"net/http"

	authdto "github.com/codec404/chat-service/dto/auth"
	"github.com/codec404/chat-service/pkg/errorhandler"
	externalerror "github.com/codec404/chat-service/pkg/external_error"
	"github.com/codec404/chat-service/pkg/render"
	authsvc "github.com/codec404/chat-service/service/auth"
)

type Handler struct {
	svc authsvc.Service
}

func NewHandler(svc authsvc.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req authdto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorhandler.Respond(w, r, externalerror.BadRequest("invalid request body"))
		return
	}

	resp, err := h.svc.Register(r.Context(), req)
	if err != nil {
		errorhandler.Respond(w, r, err)
		return
	}

	render.JSONResponse(w, r, http.StatusCreated, resp)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req authdto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorhandler.Respond(w, r, externalerror.BadRequest("invalid request body"))
		return
	}

	resp, err := h.svc.Login(r.Context(), req)
	if err != nil {
		errorhandler.Respond(w, r, err)
		return
	}

	render.JSONResponse(w, r, http.StatusOK, resp)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req authdto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorhandler.Respond(w, r, externalerror.BadRequest("invalid request body"))
		return
	}

	resp, err := h.svc.Refresh(r.Context(), req)
	if err != nil {
		errorhandler.Respond(w, r, err)
		return
	}

	render.JSONResponse(w, r, http.StatusOK, resp)
}

// Logout revokes the refresh token supplied in the request body.
// The refresh token acts as proof of identity, so no JWT is required.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req authdto.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		errorhandler.Respond(w, r, externalerror.BadRequest("refresh_token is required"))
		return
	}

	if err := h.svc.Logout(r.Context(), req.RefreshToken); err != nil {
		errorhandler.Respond(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
