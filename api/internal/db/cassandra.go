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
	// Use a simple cluster config for keyspace creation (no token-aware policy)
	initCluster := gocql.NewCluster(hosts...)
	initCluster.Consistency = gocql.Quorum
	initCluster.ProtoVersion = 4
	initCluster.ConnectTimeout = time.Second * 10
	initCluster.Timeout = time.Second * 10

	session, err := initCluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Cassandra: %w", err)
	}

	log.Printf("Connected to Cassandra cluster at %v", hosts)

	createKeyspace := fmt.Sprintf(
		`CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`,
		keyspace,
	)
	if err := session.Query(createKeyspace).Exec(); err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to create keyspace: %w", err)
	}

	log.Printf("Keyspace %s created/verified", keyspace)
	session.Close()

	// Create the main session with full config including token-aware policy
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.Quorum
	cluster.ProtoVersion = 4
	cluster.ConnectTimeout = time.Second * 10
	cluster.Timeout = time.Second * 10
	cluster.NumConns = 4
	cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())
	cluster.RetryPolicy = &gocql.ExponentialBackoffRetryPolicy{NumRetries: 3}

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

		// Secondary indexes for efficient lookups (materialized views require Cassandra config)
		// indexes on users(username) and users(email) already created above
		// Additional index for member lookups by user_id
		`CREATE INDEX IF NOT EXISTS ON members (user_id)`,

		`CREATE TABLE IF NOT EXISTS invites (
			id uuid PRIMARY KEY,
			server_id uuid,
			code text,
			created_by uuid,
			max_uses int,
			uses int,
			expires_at timestamp,
			created_at timestamp
		)`,

		`CREATE INDEX IF NOT EXISTS ON invites (code)`,
		`CREATE INDEX IF NOT EXISTS ON invites (server_id)`,
	}

	for _, query := range queries {
		if err := db.Session.Query(query).Exec(); err != nil {
			return fmt.Errorf("failed to execute schema query: %w", err)
		}
	}

	log.Println("Schema initialized successfully")
	return nil
}
