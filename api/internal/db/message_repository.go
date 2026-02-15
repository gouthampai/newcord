package db

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"newcord/api/internal/models"
)

type MessageRepository struct {
	db *CassandraDB
}

func NewMessageRepository(db *CassandraDB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(message *models.Message) error {
	message.ID = gocql.TimeUUID()
	message.CreatedAt = time.Now()

	query := `INSERT INTO messages (id, channel_id, user_id, content, type,
		attachments, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`

	return r.db.Session.Query(query,
		message.ID, message.ChannelID, message.UserID, message.Content,
		message.Type, message.Attachments, message.CreatedAt,
	).Exec()
}

func (r *MessageRepository) GetByID(channelID, messageID gocql.UUID) (*models.Message, error) {
	var message models.Message
	query := `SELECT id, channel_id, user_id, content, type, attachments,
		edited_at, created_at FROM messages WHERE channel_id = ? AND id = ? ALLOW FILTERING`

	err := r.db.Session.Query(query, channelID, messageID).Scan(
		&message.ID, &message.ChannelID, &message.UserID, &message.Content,
		&message.Type, &message.Attachments, &message.EditedAt, &message.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("message not found: %w", err)
	}

	return &message, nil
}

func (r *MessageRepository) GetByChannel(channelID gocql.UUID, limit int) ([]*models.Message, error) {
	var messages []*models.Message

	query := `SELECT id, channel_id, user_id, content, type, attachments,
		edited_at, created_at FROM messages WHERE channel_id = ? LIMIT ?`

	iter := r.db.Session.Query(query, channelID, limit).Iter()
	defer iter.Close()

	for {
		message := &models.Message{}
		if !iter.Scan(
			&message.ID, &message.ChannelID, &message.UserID, &message.Content,
			&message.Type, &message.Attachments, &message.EditedAt, &message.CreatedAt,
		) {
			break
		}
		messages = append(messages, message)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("error reading messages: %w", err)
	}

	return messages, nil
}

func (r *MessageRepository) Update(message *models.Message) error {
	now := time.Now()
	message.EditedAt = &now

	query := `UPDATE messages SET content = ?, edited_at = ?
		WHERE channel_id = ? AND created_at = ? AND id = ?`

	return r.db.Session.Query(query,
		message.Content, message.EditedAt,
		message.ChannelID, message.CreatedAt, message.ID,
	).Exec()
}

func (r *MessageRepository) Delete(channelID gocql.UUID, createdAt time.Time, messageID gocql.UUID) error {
	query := `DELETE FROM messages WHERE channel_id = ? AND created_at = ? AND id = ?`
	return r.db.Session.Query(query, channelID, createdAt, messageID).Exec()
}
