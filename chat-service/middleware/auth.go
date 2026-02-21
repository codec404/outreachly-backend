package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/codec404/chat-service/pkg/errorhandler"
	externalerror "github.com/codec404/chat-service/pkg/external_error"
)

type jwtClaims struct {
	Email string   `json:"email"`
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

// Authenticate returns a middleware that validates the JWT in the Authorization
// header and stores the user's ID, email, and roles in the request context.
func Authenticate(secret string) func(http.Handler) http.Handler {
	key := []byte(secret)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				errorhandler.Respond(w, r, externalerror.Unauthorized("missing authorization header"))
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				errorhandler.Respond(w, r, externalerror.Unauthorized("invalid authorization header format"))
				return
			}

			c := &jwtClaims{}
			token, err := jwt.ParseWithClaims(parts[1], c, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("unexpected signing method")
				}
				return key, nil
			})
			if err != nil || !token.Valid {
				errorhandler.Respond(w, r, externalerror.Unauthorized("invalid or expired token"))
				return
			}

			ctx := SetUserContext(r.Context(), c.Subject, c.Email, c.Roles)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole rejects the request with 403 if the authenticated user does not
// hold at least one of the specified roles. Must be chained after Authenticate.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, role := range RolesFromContext(r.Context()) {
				if _, ok := allowed[role]; ok {
					next.ServeHTTP(w, r)
					return
				}
			}
			errorhandler.Respond(w, r, externalerror.Forbidden("insufficient permissions"))
		})
	}
}
