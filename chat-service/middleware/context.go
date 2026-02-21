package middleware

import "context"

type contextKey string

const (
	userIDKey    contextKey = "user_id"
	userEmailKey contextKey = "user_email"
	userRolesKey contextKey = "user_roles"
)

// SetUserContext stores the authenticated user's ID, email, and roles in the context.
// Called by Authenticate after successful JWT verification.
func SetUserContext(ctx context.Context, userID, email string, roles []string) context.Context {
	ctx = context.WithValue(ctx, userIDKey, userID)
	ctx = context.WithValue(ctx, userEmailKey, email)
	ctx = context.WithValue(ctx, userRolesKey, roles)
	return ctx
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok
}

func UserEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(userEmailKey).(string)
	return email, ok
}

func RolesFromContext(ctx context.Context) []string {
	roles, _ := ctx.Value(userRolesKey).([]string)
	return roles
}
