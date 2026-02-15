package models

import (
	"time"

	"github.com/gocql/gocql"
)

type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeFile  MessageType = "file"
	MessageTypeAudio MessageType = "audio"
)

type Message struct {
	ID          gocql.UUID  `json:"id"`
	ChannelID   gocql.UUID  `json:"channel_id"`
	UserID      gocql.UUID  `json:"user_id"`
	Content     string      `json:"content"`
	Type        MessageType `json:"type"`
	Attachments []string    `json:"attachments"`
	EditedAt    *time.Time  `json:"edited_at,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
}

type MessageReaction struct {
	MessageID gocql.UUID `json:"message_id"`
	UserID    gocql.UUID `json:"user_id"`
	Emoji     string     `json:"emoji"`
	CreatedAt time.Time  `json:"created_at"`
}
