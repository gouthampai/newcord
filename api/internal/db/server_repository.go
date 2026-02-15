package db

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"newcord/api/internal/models"
)

type ServerRepository struct {
	db *CassandraDB
}

func NewServerRepository(db *CassandraDB) *ServerRepository {
	return &ServerRepository{db: db}
}

func (r *ServerRepository) Create(server *models.Server) error {
	server.ID = gocql.TimeUUID()
	server.CreatedAt = time.Now()
	server.UpdatedAt = time.Now()

	query := `INSERT INTO servers (id, name, description, icon_url, owner_id,
		created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`

	return r.db.Session.Query(query,
		server.ID, server.Name, server.Description, server.IconURL,
		server.OwnerID, server.CreatedAt, server.UpdatedAt,
	).Exec()
}

func (r *ServerRepository) GetByID(id gocql.UUID) (*models.Server, error) {
	var server models.Server
	query := `SELECT id, name, description, icon_url, owner_id, created_at,
		updated_at FROM servers WHERE id = ?`

	err := r.db.Session.Query(query, id).Scan(
		&server.ID, &server.Name, &server.Description, &server.IconURL,
		&server.OwnerID, &server.CreatedAt, &server.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}

	return &server, nil
}

func (r *ServerRepository) Update(server *models.Server) error {
	server.UpdatedAt = time.Now()

	query := `UPDATE servers SET name = ?, description = ?, icon_url = ?,
		updated_at = ? WHERE id = ?`

	return r.db.Session.Query(query,
		server.Name, server.Description, server.IconURL,
		server.UpdatedAt, server.ID,
	).Exec()
}

func (r *ServerRepository) Delete(id gocql.UUID) error {
	query := `DELETE FROM servers WHERE id = ?`
	return r.db.Session.Query(query, id).Exec()
}

func (r *ServerRepository) AddMember(member *models.Member) error {
	member.JoinedAt = time.Now()
	member.UpdatedAt = time.Now()

	query := `INSERT INTO members (server_id, user_id, nickname, role,
		joined_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`

	return r.db.Session.Query(query,
		member.ServerID, member.UserID, member.Nickname,
		member.Role, member.JoinedAt, member.UpdatedAt,
	).Exec()
}

func (r *ServerRepository) GetMembers(serverID gocql.UUID) ([]*models.Member, error) {
	var members []*models.Member

	query := `SELECT server_id, user_id, nickname, role, joined_at, updated_at
		FROM members WHERE server_id = ?`

	iter := r.db.Session.Query(query, serverID).Iter()
	defer iter.Close()

	for {
		member := &models.Member{}
		if !iter.Scan(
			&member.ServerID, &member.UserID, &member.Nickname,
			&member.Role, &member.JoinedAt, &member.UpdatedAt,
		) {
			break
		}
		members = append(members, member)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("error reading members: %w", err)
	}

	return members, nil
}

func (r *ServerRepository) RemoveMember(serverID, userID gocql.UUID) error {
	query := `DELETE FROM members WHERE server_id = ? AND user_id = ?`
	return r.db.Session.Query(query, serverID, userID).Exec()
}
