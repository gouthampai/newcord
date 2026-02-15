# Newcord API

A self-hosted Discord-like chat application API built with Go and CassandraDB.

## Features

- User authentication with JWT
- Server/guild management
- Text channels
- Real-time messaging via WebSockets
- Message history
- User presence tracking

## Tech Stack

- **Language**: Go 1.21+
- **Database**: Apache Cassandra
- **Real-time**: WebSockets
- **Authentication**: JWT

## Project Structure

```
api/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── db/              # Database connection and repositories
│   ├── handlers/        # HTTP handlers
│   ├── middleware/      # HTTP middleware
│   ├── models/          # Data models
│   └── websocket/       # WebSocket handling
├── pkg/
│   ├── config/          # Configuration management
│   └── utils/           # Utility functions
└── go.mod
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Kubernetes cluster (Docker Desktop with Kubernetes, Minikube, or similar)
- kubectl CLI tool

### Installation

1. Clone the repository
2. Copy the environment file:
   ```bash
   cp .env.example .env
   ```

3. Deploy Cassandra to Kubernetes:
   ```bash
   make k8s-deploy
   ```

4. Wait for Cassandra to be ready:
   ```bash
   make k8s-status
   # Wait until cassandra-0 pod is Running and Ready (1/1)
   ```

5. (Optional) For local development, port-forward Cassandra:
   ```bash
   make k8s-port-forward
   # Or connect directly using the NodePort at localhost:30042
   ```

6. Install dependencies:
   ```bash
   make install-deps
   ```

7. Run the server:
   ```bash
   make run
   ```

The server will start on `http://localhost:8080`

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register a new user
- `POST /api/v1/auth/login` - Login user

### Users
- `GET /api/v1/users/{id}` - Get user by ID
- `PUT /api/v1/users/{id}` - Update user
- `DELETE /api/v1/users/{id}` - Delete user

### Servers
- `POST /api/v1/servers` - Create a server
- `GET /api/v1/servers/{id}` - Get server by ID
- `PUT /api/v1/servers/{id}` - Update server
- `DELETE /api/v1/servers/{id}` - Delete server
- `GET /api/v1/servers/{id}/members` - Get server members
- `POST /api/v1/servers/{id}/members` - Add member to server

### Channels
- `POST /api/v1/channels` - Create a channel
- `GET /api/v1/channels/{id}` - Get channel by ID
- `PUT /api/v1/channels/{id}` - Update channel
- `DELETE /api/v1/channels/{id}` - Delete channel
- `GET /api/v1/servers/{server_id}/channels` - Get server channels

### Messages
- `POST /api/v1/channels/{channel_id}/messages` - Send a message
- `GET /api/v1/channels/{channel_id}/messages` - Get messages (with ?limit=N)
- `PUT /api/v1/channels/{channel_id}/messages/{message_id}` - Edit message
- `DELETE /api/v1/channels/{channel_id}/messages/{message_id}` - Delete message

### WebSocket
- `WS /api/v1/ws/{server_id}` - Connect to server WebSocket

## Database Schema

The application automatically creates the following Cassandra tables:
- `users` - User accounts
- `servers` - Discord-like servers/guilds
- `members` - Server membership
- `channels` - Text/voice channels
- `messages` - Chat messages
- `direct_messages` - DM conversations
- `user_presence` - User online status

## Development

### Building
```bash
make build
```

### Running Tests
```bash
make test
```

### Cleaning Build Artifacts
```bash
make clean
```

## Kubernetes Management

### Deploy Cassandra
```bash
make k8s-deploy
```

### Check Status
```bash
make k8s-status
```

### View Logs
```bash
make k8s-logs
```

### Connect to Cassandra (cqlsh)
```bash
make k8s-connect
```

### Port Forward (for local development)
```bash
make k8s-port-forward
# Cassandra will be available at localhost:9042
```

### Delete Deployment
```bash
make k8s-delete
```

## Environment Variables

- `PORT` - Server port (default: 8080)
- `CASSANDRA_HOSTS` - Comma-separated Cassandra hosts (default: localhost)
- `CASSANDRA_KEYSPACE` - Cassandra keyspace (default: newcord)
- `JWT_SECRET` - Secret key for JWT signing (change in production!)

## WebSocket Message Format

```json
{
  "type": "message",
  "channel_id": "uuid",
  "server_id": "uuid",
  "data": {
    "content": "Hello!",
    "type": "text"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## License

MIT
