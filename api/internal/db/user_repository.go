package db

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"newcord/api/internal/models"
)

type UserRepository struct {
	db *CassandraDB
}

func NewUserRepository(db *CassandraDB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	user.ID = gocql.TimeUUID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `INSERT INTO users (id, username, email, password_hash, display_name,
		avatar_url, status, bio, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	return r.db.Session.Query(query,
		user.ID, user.Username, user.Email, user.PasswordHash,
		user.DisplayName, user.AvatarURL, user.Status, user.Bio,
		user.CreatedAt, user.UpdatedAt,
	).Exec()
}

func (r *UserRepository) GetByID(id gocql.UUID) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, email, password_hash, display_name, avatar_url,
		status, bio, created_at, updated_at FROM users WHERE id = ?`

	err := r.db.Session.Query(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.DisplayName, &user.AvatarURL, &user.Status, &user.Bio,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, email, password_hash, display_name, avatar_url,
		status, bio, created_at, updated_at FROM users WHERE username = ? ALLOW FILTERING`

	err := r.db.Session.Query(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.DisplayName, &user.AvatarURL, &user.Status, &user.Bio,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, email, password_hash, display_name, avatar_url,
		status, bio, created_at, updated_at FROM users WHERE email = ? ALLOW FILTERING`

	err := r.db.Session.Query(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.DisplayName, &user.AvatarURL, &user.Status, &user.Bio,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) Update(user *models.User) error {
	user.UpdatedAt = time.Now()

	query := `UPDATE users SET username = ?, email = ?, display_name = ?,
		avatar_url = ?, status = ?, bio = ?, updated_at = ? WHERE id = ?`

	return r.db.Session.Query(query,
		user.Username, user.Email, user.DisplayName,
		user.AvatarURL, user.Status, user.Bio, user.UpdatedAt, user.ID,
	).Exec()
}

func (r *UserRepository) Delete(id gocql.UUID) error {
	query := `DELETE FROM users WHERE id = ?`
	return r.db.Session.Query(query, id).Exec()
}
