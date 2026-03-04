package auth

import "github.com/codec404/chat-service/utils"

const (
	oauthStateCookieName = "oauth_state"
	googleUserInfoURL    = "https://www.googleapis.com/oauth2/v3/userinfo"
)

// generateOAuthState returns a cryptographically random hex string used as the
// CSRF state parameter in the OAuth2 flow.
func generateOAuthState() (string, error) {
	return utils.GenerateToken()
}
