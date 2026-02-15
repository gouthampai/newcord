package models

import (
	"time"

	"github.com/gocql/gocql"
)

type Server struct {
	ID          gocql.UUID `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	IconURL     string     `json:"icon_url"`
	OwnerID     gocql.UUID `json:"owner_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Member struct {
	ServerID  gocql.UUID `json:"server_id"`
	UserID    gocql.UUID `json:"user_id"`
	Nickname  string     `json:"nickname"`
	Role      string     `json:"role"` // owner, admin, moderator, member
	JoinedAt  time.Time  `json:"joined_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
