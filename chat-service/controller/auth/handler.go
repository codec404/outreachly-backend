package auth

import (
	"encoding/json"
	"net/http"

	"golang.org/x/oauth2"

	authdto "github.com/codec404/chat-service/dto/auth"
	"github.com/codec404/chat-service/pkg/errorhandler"
	externalerror "github.com/codec404/chat-service/pkg/external_error"
	log "github.com/codec404/chat-service/pkg/logger"
	"github.com/codec404/chat-service/pkg/render"
	authsvc "github.com/codec404/chat-service/service/auth"
)

type Handler struct {
	svc           authsvc.Service
	oauthConfig   *oauth2.Config
	secureCookies bool
	stateCookieMaxAge int
}

func NewHandler(svc authsvc.Service, oauthConfig *oauth2.Config, secureCookies bool, stateCookieMaxAge int) *Handler {
	return &Handler{
		svc:               svc,
		oauthConfig:       oauthConfig,
		secureCookies:     secureCookies,
		stateCookieMaxAge: stateCookieMaxAge,
	}
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

// GoogleLogin redirects the browser to Google's OAuth consent page.
// A random state token is stored in a short-lived HttpOnly cookie for CSRF protection.
func (h *Handler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := generateOAuthState()
	if err != nil {
		errorhandler.Respond(w, r, externalerror.Internal())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    state,
		MaxAge:   h.stateCookieMaxAge,
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	http.Redirect(w, r, h.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOnline), http.StatusTemporaryRedirect)
}

// GoogleCallback handles the redirect from Google, exchanges the code for tokens,
// fetches user info, then issues our own JWT + refresh token pair.
func (h *Handler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// 1. Validate CSRF state.
	stateCookie, err := r.Cookie(oauthStateCookieName)
	if err != nil || stateCookie.Value == "" {
		errorhandler.Respond(w, r, externalerror.BadRequest("missing oauth state"))
		return
	}
	// Clear the state cookie immediately regardless of outcome.
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
	})
	if r.URL.Query().Get("state") != stateCookie.Value {
		errorhandler.Respond(w, r, externalerror.BadRequest("invalid oauth state"))
		return
	}

	// 2. Check if Google returned an error (e.g. user denied consent).
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		errorhandler.Respond(w, r, externalerror.BadRequest("google oauth denied: "+errParam))
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		errorhandler.Respond(w, r, externalerror.BadRequest("missing oauth code"))
		return
	}

	// 3. Exchange the authorization code for a Google access token.
	googleToken, err := h.oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "google oauth: code exchange failed: %v", err)
		errorhandler.Respond(w, r, externalerror.Internal())
		return
	}

	// 4. Fetch user info from Google's userinfo endpoint.
	client := h.oauthConfig.Client(r.Context(), googleToken)
	resp, err := client.Get(googleUserInfoURL)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "google oauth: userinfo fetch failed: %v", err)
		errorhandler.Respond(w, r, externalerror.Internal())
		return
	}
	defer resp.Body.Close()

	var googleUser struct {
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		log.ErrorfWithContext(r.Context(), "google oauth: decode userinfo failed: %v", err)
		errorhandler.Respond(w, r, externalerror.Internal())
		return
	}
	if googleUser.Sub == "" || googleUser.Email == "" {
		log.ErrorfWithContext(r.Context(), "google oauth: userinfo missing required fields")
		errorhandler.Respond(w, r, externalerror.Internal())
		return
	}

	// 5. Delegate to the service — login or create + link.
	authResp, err := h.svc.LoginOrRegisterWithGoogle(r.Context(), authsvc.GoogleUserInfo{
		Sub:       googleUser.Sub,
		Email:     googleUser.Email,
		Name:      googleUser.Name,
		AvatarURL: googleUser.Picture,
	})
	if err != nil {
		errorhandler.Respond(w, r, err)
		return
	}

	render.JSONResponse(w, r, http.StatusOK, authResp)
}
