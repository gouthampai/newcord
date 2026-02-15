package db

import (
	"fmt"
	"log"
	"time"

	"github.com/gocql/gocql"
)

type CassandraDB struct {
	Session *gocql.Session
}

func NewCassandraDB(hosts []string, keyspace string) (*CassandraDB, error) {
	// First connect without keyspace to create it if needed
	cluster := gocql.NewCluster(hosts...)
	cluster.Consistency = gocql.Quorum
	cluster.ProtoVersion = 4
	cluster.ConnectTimeout = time.Second * 10
	cluster.Timeout = time.Second * 10

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Cassandra: %w", err)
	}

	log.Printf("Connected to Cassandra cluster at %v", hosts)

	// Create keyspace if it doesn't exist
	createKeyspace := fmt.Sprintf(
		`CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`,
		keyspace,
	)
	if err := session.Query(createKeyspace).Exec(); err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to create keyspace: %w", err)
	}

	log.Printf("Keyspace %s created/verified", keyspace)

	// Close the initial session
	session.Close()

	// Reconnect with the keyspace set
	cluster.Keyspace = keyspace
	session, err = cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to keyspace: %w", err)
	}

	log.Printf("Connected to keyspace: %s", keyspace)

	return &CassandraDB{Session: session}, nil
}

func (db *CassandraDB) Close() {
	if db.Session != nil {
		db.Session.Close()
	}
}

func (db *CassandraDB) InitSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id uuid PRIMARY KEY,
			username text,
			email text,
			password_hash text,
			display_name text,
			avatar_url text,
			status text,
			bio text,
			created_at timestamp,
			updated_at timestamp
		)`,

		`CREATE INDEX IF NOT EXISTS ON users (username)`,
		`CREATE INDEX IF NOT EXISTS ON users (email)`,

		`CREATE TABLE IF NOT EXISTS servers (
			id uuid PRIMARY KEY,
			name text,
			description text,
			icon_url text,
			owner_id uuid,
			created_at timestamp,
			updated_at timestamp
		)`,

		`CREATE TABLE IF NOT EXISTS members (
			server_id uuid,
			user_id uuid,
			nickname text,
			role text,
			joined_at timestamp,
			updated_at timestamp,
			PRIMARY KEY (server_id, user_id)
		)`,

		`CREATE TABLE IF NOT EXISTS channels (
			id uuid PRIMARY KEY,
			server_id uuid,
			name text,
			description text,
			type text,
			position int,
			created_at timestamp,
			updated_at timestamp
		)`,

		`CREATE INDEX IF NOT EXISTS ON channels (server_id)`,

		`CREATE TABLE IF NOT EXISTS messages (
			id uuid,
			channel_id uuid,
			user_id uuid,
			content text,
			type text,
			attachments list<text>,
			edited_at timestamp,
			created_at timestamp,
			PRIMARY KEY (channel_id, created_at, id)
		) WITH CLUSTERING ORDER BY (created_at DESC)`,

		`CREATE TABLE IF NOT EXISTS direct_messages (
			id uuid PRIMARY KEY,
			participants list<uuid>,
			created_at timestamp
		)`,

		`CREATE TABLE IF NOT EXISTS user_presence (
			user_id uuid PRIMARY KEY,
			status text,
			custom_status text,
			last_seen_at timestamp
		)`,
	}

	for _, query := range queries {
		if err := db.Session.Query(query).Exec(); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	log.Println("Schema initialized successfully")
	return nil
}
