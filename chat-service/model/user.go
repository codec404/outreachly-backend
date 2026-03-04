package model

import "time"

type User struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string
	AvatarURL    string
	IsActive     bool
	IsBlocked    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
