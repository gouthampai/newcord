package db

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"newcord/api/internal/models"
)

type ChannelRepository struct {
	db *CassandraDB
}

func NewChannelRepository(db *CassandraDB) *ChannelRepository {
	return &ChannelRepository{db: db}
}

func (r *ChannelRepository) Create(channel *models.Channel) error {
	channel.ID = gocql.TimeUUID()
	channel.CreatedAt = time.Now()
	channel.UpdatedAt = time.Now()

	query := `INSERT INTO channels (id, server_id, name, description, type,
		position, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	return r.db.Session.Query(query,
		channel.ID, channel.ServerID, channel.Name, channel.Description,
		channel.Type, channel.Position, channel.CreatedAt, channel.UpdatedAt,
	).Exec()
}

func (r *ChannelRepository) GetByID(id gocql.UUID) (*models.Channel, error) {
	var channel models.Channel
	query := `SELECT id, server_id, name, description, type, position,
		created_at, updated_at FROM channels WHERE id = ?`

	err := r.db.Session.Query(query, id).Scan(
		&channel.ID, &channel.ServerID, &channel.Name, &channel.Description,
		&channel.Type, &channel.Position, &channel.CreatedAt, &channel.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("channel not found: %w", err)
	}

	return &channel, nil
}

func (r *ChannelRepository) GetByServer(serverID gocql.UUID) ([]*models.Channel, error) {
	var channels []*models.Channel

	query := `SELECT id, server_id, name, description, type, position,
		created_at, updated_at FROM channels WHERE server_id = ? ALLOW FILTERING`

	iter := r.db.Session.Query(query, serverID).Iter()
	defer iter.Close()

	for {
		channel := &models.Channel{}
		if !iter.Scan(
			&channel.ID, &channel.ServerID, &channel.Name, &channel.Description,
			&channel.Type, &channel.Position, &channel.CreatedAt, &channel.UpdatedAt,
		) {
			break
		}
		channels = append(channels, channel)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("error reading channels: %w", err)
	}

	return channels, nil
}

func (r *ChannelRepository) Update(channel *models.Channel) error {
	channel.UpdatedAt = time.Now()

	query := `UPDATE channels SET name = ?, description = ?, position = ?,
		updated_at = ? WHERE id = ?`

	return r.db.Session.Query(query,
		channel.Name, channel.Description, channel.Position,
		channel.UpdatedAt, channel.ID,
	).Exec()
}

func (r *ChannelRepository) Delete(id gocql.UUID) error {
	query := `DELETE FROM channels WHERE id = ?`
	return r.db.Session.Query(query, id).Exec()
}
