package authsvc

import (
	"context"

	authdto "github.com/codec404/chat-service/dto/auth"
)

// GoogleUserInfo holds the user data fetched from Google's userinfo endpoint.
// It is internal to the service — never decoded from an HTTP request body.
type GoogleUserInfo struct {
	Sub       string // Google's unique user ID ("sub" claim)
	Email     string
	Name      string
	AvatarURL string // "picture" field from the userinfo endpoint
}

type Service interface {
	Register(ctx context.Context, req authdto.RegisterRequest) (*authdto.AuthResponse, error)
	Login(ctx context.Context, req authdto.LoginRequest) (*authdto.AuthResponse, error)
	Refresh(ctx context.Context, req authdto.RefreshRequest) (*authdto.AuthResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	LoginOrRegisterWithGoogle(ctx context.Context, info GoogleUserInfo) (*authdto.AuthResponse, error)
}
