package authsvc

import (
	"context"

	authdto "github.com/codec404/chat-service/dto/auth"
)

type Service interface {
	Register(ctx context.Context, req authdto.RegisterRequest) (*authdto.AuthResponse, error)
	Login(ctx context.Context, req authdto.LoginRequest) (*authdto.AuthResponse, error)
	Refresh(ctx context.Context, req authdto.RefreshRequest) (*authdto.AuthResponse, error)
	Logout(ctx context.Context, refreshToken string) error
}
