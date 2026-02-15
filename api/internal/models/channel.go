package models

import (
	"time"

	"github.com/gocql/gocql"
)

type ChannelType string

const (
	ChannelTypeText  ChannelType = "text"
	ChannelTypeVoice ChannelType = "voice"
	ChannelTypeDM    ChannelType = "dm"
)

type Channel struct {
	ID          gocql.UUID  `json:"id"`
	ServerID    gocql.UUID  `json:"server_id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Type        ChannelType `json:"type"`
	Position    int         `json:"position"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type DirectMessage struct {
	ID           gocql.UUID   `json:"id"`
	Participants []gocql.UUID `json:"participants"`
	CreatedAt    time.Time    `json:"created_at"`
}
