package db

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"newcord/api/internal/models"
)

type InviteRepository struct {
	db *CassandraDB
}

func NewInviteRepository(db *CassandraDB) *InviteRepository {
	return &InviteRepository{db: db}
}

func (r *InviteRepository) Create(invite *models.Invite) error {
	invite.ID = gocql.TimeUUID()
	invite.CreatedAt = time.Now()

	query := `INSERT INTO invites (id, server_id, code, created_by, max_uses, uses, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	return r.db.Session.Query(query,
		invite.ID, invite.ServerID, invite.Code, invite.CreatedBy,
		invite.MaxUses, invite.Uses, invite.ExpiresAt, invite.CreatedAt,
	).Exec()
}

func (r *InviteRepository) GetByCode(code string) (*models.Invite, error) {
	var invite models.Invite
	query := `SELECT id, server_id, code, created_by, max_uses, uses, expires_at, created_at
		FROM invites WHERE code = ?`

	err := r.db.Session.Query(query, code).Scan(
		&invite.ID, &invite.ServerID, &invite.Code, &invite.CreatedBy,
		&invite.MaxUses, &invite.Uses, &invite.ExpiresAt, &invite.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("invite not found: %w", err)
	}
	return &invite, nil
}

func (r *InviteRepository) GetByServer(serverID gocql.UUID) ([]*models.Invite, error) {
	var invites []*models.Invite

	query := `SELECT id, server_id, code, created_by, max_uses, uses, expires_at, created_at
		FROM invites WHERE server_id = ?`

	iter := r.db.Session.Query(query, serverID).Iter()
	for {
		invite := &models.Invite{}
		if !iter.Scan(
			&invite.ID, &invite.ServerID, &invite.Code, &invite.CreatedBy,
			&invite.MaxUses, &invite.Uses, &invite.ExpiresAt, &invite.CreatedAt,
		) {
			break
		}
		invites = append(invites, invite)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("error reading invites: %w", err)
	}
	return invites, nil
}

func (r *InviteRepository) Delete(id gocql.UUID) error {
	query := `DELETE FROM invites WHERE id = ?`
	return r.db.Session.Query(query, id).Exec()
}

func (r *InviteRepository) IncrementUses(id gocql.UUID, currentUses int) error {
	query := `UPDATE invites SET uses = ? WHERE id = ?`
	return r.db.Session.Query(query, currentUses+1, id).Exec()
}
