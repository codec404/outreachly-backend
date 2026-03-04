package authsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"

	authdto "github.com/codec404/chat-service/dto/auth"
	"github.com/codec404/chat-service/model"
	externalerror "github.com/codec404/chat-service/pkg/external_error"
	log "github.com/codec404/chat-service/pkg/logger"
	oauthrepo "github.com/codec404/chat-service/repository/oauth"
	tokenrepo "github.com/codec404/chat-service/repository/token"
	userrepo "github.com/codec404/chat-service/repository/user"
	"github.com/codec404/chat-service/utils"
)

type service struct {
	userRepo         userrepo.Repository
	tokenRepo        tokenrepo.Repository
	oauthRepo        oauthrepo.Repository
	oauth2Config     *oauth2.Config
	jwtSecret        []byte
	accessExpiryMin  int
	refreshExpiryDay int
}

func New(
	userRepo userrepo.Repository,
	tokenRepo tokenrepo.Repository,
	oauthRepo oauthrepo.Repository,
	oauth2Config *oauth2.Config,
	jwtSecret string,
	accessExpiryMin int,
	refreshExpiryDay int,
) Service {
	return &service{
		userRepo:         userRepo,
		tokenRepo:        tokenRepo,
		oauthRepo:        oauthRepo,
		oauth2Config:     oauth2Config,
		jwtSecret:        []byte(jwtSecret),
		accessExpiryMin:  accessExpiryMin,
		refreshExpiryDay: refreshExpiryDay,
	}
}

type jwtClaims struct {
	Email string   `json:"email"`
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

// Register creates a new user with the "user" role and returns tokens.
func (s *service) Register(ctx context.Context, req authdto.RegisterRequest) (*authdto.AuthResponse, error) {
	if err := validateRegister(req); err != nil {
		log.WarnfWithContext(ctx, "audit: register rejected email=%s reason=%s", req.Email, err.Error())
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("authsvc.Register hash password: %w", err)
	}

	user, err := s.userRepo.CreateWithRole(ctx, req.Name, req.Email, string(hash), "user")
	if err != nil {
		log.WarnfWithContext(ctx, "audit: register failed email=%s err=%v", req.Email, err)
		return nil, err
	}

	roles, err := s.userRepo.GetRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	log.InfofWithContext(ctx, "audit: register success email=%s user_id=%s", user.Email, user.ID)
	return s.issueTokens(ctx, user, roles)
}

// Login validates credentials and returns tokens.
func (s *service) Login(ctx context.Context, req authdto.LoginRequest) (*authdto.AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, externalerror.BadRequest("email and password are required")
	}

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		// Generic error prevents email enumeration.
		log.WarnfWithContext(ctx, "audit: login failed email=%s reason=user_not_found", req.Email)
		return nil, externalerror.Unauthorized("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		log.WarnfWithContext(ctx, "audit: login failed email=%s reason=bad_password", req.Email)
		return nil, externalerror.Unauthorized("invalid credentials")
	}

	if user.IsBlocked {
		log.WarnfWithContext(ctx, "audit: login denied email=%s user_id=%s reason=blocked", user.Email, user.ID)
		return nil, externalerror.Forbidden("account is blocked")
	}

	roles, err := s.userRepo.GetRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	log.InfofWithContext(ctx, "audit: login success email=%s user_id=%s", user.Email, user.ID)
	return s.issueTokens(ctx, user, roles)
}

// Refresh rotates the refresh token and returns a new token pair.
// FindAndRevokeByHash atomically validates + revokes in one SQL statement,
// so concurrent calls with the same token cannot both succeed.
// The IsBlocked check after is intentionally post-revocation: the token is
// already consumed, so a blocked user cannot retry with the same token.
func (s *service) Refresh(ctx context.Context, req authdto.RefreshRequest) (*authdto.AuthResponse, error) {
	if req.RefreshToken == "" {
		return nil, externalerror.BadRequest("refresh_token is required")
	}

	hash := utils.HashToken(req.RefreshToken)
	stored, err := s.tokenRepo.FindAndRevokeByHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, stored.UserID)
	if err != nil {
		return nil, err
	}
	if user.IsBlocked {
		log.WarnfWithContext(ctx, "audit: refresh denied user_id=%s reason=blocked", stored.UserID)
		return nil, externalerror.Forbidden("account is blocked")
	}

	roles, err := s.userRepo.GetRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	log.InfofWithContext(ctx, "audit: refresh success user_id=%s", user.ID)
	return s.issueTokens(ctx, user, roles)
}

// Logout revokes the provided refresh token (no-op if token not found).
func (s *service) Logout(ctx context.Context, refreshToken string) error {
	hash := utils.HashToken(refreshToken)
	stored, err := s.tokenRepo.FindByHash(ctx, hash)
	if err != nil {
		// Token not found — already logged out; treat as success.
		return nil
	}
	if err := s.tokenRepo.Revoke(ctx, stored.ID); err != nil {
		return err
	}
	log.InfofWithContext(ctx, "audit: logout success user_id=%s", stored.UserID)
	return nil
}

// LoginOrRegisterWithGoogle either logs in an existing Google-linked user,
// links a Google account to an existing local user with the same email,
// or creates a brand-new user — then issues tokens.
func (s *service) LoginOrRegisterWithGoogle(ctx context.Context, info GoogleUserInfo) (*authdto.AuthResponse, error) {
	// 1. Check if this Google account is already linked to a local user.
	existing, err := s.oauthRepo.FindByProvider(ctx, "google", info.Sub)
	if err == nil {
		user, err := s.userRepo.FindByID(ctx, existing.UserID)
		if err != nil {
			return nil, err
		}
		if user.IsBlocked {
			log.WarnfWithContext(ctx, "audit: google login denied user_id=%s reason=blocked", user.ID)
			return nil, externalerror.Forbidden("account is blocked")
		}
		if info.AvatarURL != "" && user.AvatarURL == "" {
			_ = s.userRepo.UpdateAvatarURL(ctx, user.ID, info.AvatarURL)
		}
		roles, err := s.userRepo.GetRoles(ctx, user.ID)
		if err != nil {
			return nil, err
		}
		log.InfofWithContext(ctx, "audit: google login success user_id=%s", user.ID)
		return s.issueTokens(ctx, user, roles)
	}

	// 2. No existing link — find or create a local user by email.
	user, err := s.userRepo.FindByEmail(ctx, info.Email)
	if err != nil {
		// No local user with this email — create one.
		randomHash, err := generateRandomPasswordHash()
		if err != nil {
			return nil, fmt.Errorf("authsvc.LoginOrRegisterWithGoogle: %w", err)
		}
		user, err = s.userRepo.CreateWithRole(ctx, info.Name, info.Email, randomHash, "user")
		if err != nil {
			return nil, err
		}
		log.InfofWithContext(ctx, "audit: google register success user_id=%s", user.ID)
	} else {
		log.InfofWithContext(ctx, "audit: google linked to existing user user_id=%s", user.ID)
	}

	if user.IsBlocked {
		return nil, externalerror.Forbidden("account is blocked")
	}

	if info.AvatarURL != "" && user.AvatarURL == "" {
		_ = s.userRepo.UpdateAvatarURL(ctx, user.ID, info.AvatarURL)
	}

	// 3. Link the OAuth provider row to this user.
	if _, err := s.oauthRepo.Create(ctx, oauthrepo.CreateParams{
		UserID:         user.ID,
		Provider:       "google",
		ProviderUserID: info.Sub,
		Email:          info.Email,
		Name:           info.Name,
		AvatarURL:      info.AvatarURL,
	}); err != nil {
		return nil, err
	}

	roles, err := s.userRepo.GetRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, user, roles)
}

// issueTokens mints a new JWT access token and a random refresh token.
func (s *service) issueTokens(ctx context.Context, user *model.User, roles []string) (*authdto.AuthResponse, error) {
	now := time.Now()
	accessExpiry := time.Duration(s.accessExpiryMin) * time.Minute

	c := &jwtClaims{
		Email: user.Email,
		Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(accessExpiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	accessToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("authsvc.issueTokens sign JWT: %w", err)
	}

	rawRefresh, err := utils.GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("authsvc.issueTokens generate refresh token: %w", err)
	}

	refreshExpiry := time.Duration(s.refreshExpiryDay) * 24 * time.Hour
	if _, err := s.tokenRepo.Create(ctx, user.ID, utils.HashToken(rawRefresh), now.Add(refreshExpiry)); err != nil {
		return nil, err
	}

	return &authdto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    s.accessExpiryMin * 60,
	}, nil
}

// generateRandomPasswordHash creates a bcrypt hash of a random token.
// Used for OAuth-only accounts that must satisfy the NOT NULL constraint on password_hash.
func generateRandomPasswordHash() (string, error) {
	raw, err := utils.GenerateToken()
	if err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func validateRegister(req authdto.RegisterRequest) error {
	if req.Name == "" {
		return externalerror.BadRequest("name is required")
	}
	if req.Email == "" {
		return externalerror.BadRequest("email is required")
	}
	if !utils.IsValidEmail(req.Email) {
		return externalerror.BadRequest("invalid email address")
	}
	if len(req.Password) < 8 {
		return externalerror.BadRequest("password must be at least 8 characters")
	}
	if !utils.HasUppercase(req.Password) || !utils.HasDigit(req.Password) {
		return externalerror.BadRequest("password must contain at least one uppercase letter and one digit")
	}
	return nil
}
