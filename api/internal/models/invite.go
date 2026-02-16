package models

import (
	"time"

	"github.com/gocql/gocql"
)

type Invite struct {
	ID        gocql.UUID `json:"id"`
	ServerID  gocql.UUID `json:"server_id"`
	Code      string     `json:"code"`
	CreatedBy gocql.UUID `json:"created_by"`
	MaxUses   int        `json:"max_uses"`
	Uses      int        `json:"uses"`
	ExpiresAt time.Time  `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
}
