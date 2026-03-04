package model

import "time"

type OAuthProvider struct {
	ID             string
	UserID         string
	Provider       string
	ProviderUserID string
	Email          string
	Name           string
	AvatarURL      string
	CreatedAt      time.Time
}
