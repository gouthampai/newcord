package models

import (
	"time"

	"github.com/gocql/gocql"
)

type User struct {
	ID           gocql.UUID `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	DisplayName  string     `json:"display_name"`
	AvatarURL    string     `json:"avatar_url"`
	Status       string     `json:"status"` // online, offline, away, dnd
	Bio          string     `json:"bio"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type UserPresence struct {
	UserID       gocql.UUID `json:"user_id"`
	Status       string     `json:"status"`
	CustomStatus string     `json:"custom_status"`
	LastSeenAt   time.Time  `json:"last_seen_at"`
}
