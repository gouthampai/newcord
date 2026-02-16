package websocket

import (
	"encoding/json"
	"log"

	"github.com/gocql/gocql"
)

type BroadcastMessage struct {
	ServerID gocql.UUID
	Message  []byte
}

type Hub struct {
	Clients    map[gocql.UUID]map[*Client]bool
	Broadcast  chan BroadcastMessage
	Register   chan *Client
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[gocql.UUID]map[*Client]bool),
		Broadcast:  make(chan BroadcastMessage),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// sendToServer writes a payload directly to every client's Send channel for a given server.
// Must only be called from within Run() to avoid concurrent map access.
func (h *Hub) sendToServer(serverID gocql.UUID, payload []byte) {
	if serverClients, ok := h.Clients[serverID]; ok {
		for client := range serverClients {
			select {
			case client.Send <- payload:
			default:
				close(client.Send)
				delete(serverClients, client)
				if len(serverClients) == 0 {
					delete(h.Clients, serverID)
				}
			}
		}
	}
}

// userConnectedToServer checks if a user has any active connections to a server.
// Must only be called from within Run().
func (h *Hub) userConnectedToServer(userID gocql.UUID, serverID gocql.UUID) bool {
	if serverClients, ok := h.Clients[serverID]; ok {
		for client := range serverClients {
			if client.UserID == userID {
				return true
			}
		}
	}
	return false
}

// onlineUserIDs returns deduplicated user IDs for a server.
// Must only be called from within Run().
func (h *Hub) onlineUserIDs(serverID gocql.UUID) []string {
	seen := make(map[gocql.UUID]bool)
	var ids []string
	if serverClients, ok := h.Clients[serverID]; ok {
		for client := range serverClients {
			if !seen[client.UserID] {
				seen[client.UserID] = true
				ids = append(ids, client.UserID.String())
			}
		}
	}
	return ids
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			if h.Clients[client.ServerID] == nil {
				h.Clients[client.ServerID] = make(map[*Client]bool)
			}

			// Check if user was already connected (multi-tab)
			alreadyConnected := h.userConnectedToServer(client.UserID, client.ServerID)

			h.Clients[client.ServerID][client] = true
			log.Printf("Client registered. Server: %s, Total clients: %d",
				client.ServerID, len(h.Clients[client.ServerID]))

			// If this is the user's first connection, broadcast online status
			if !alreadyConnected {
				msg, _ := json.Marshal(Message{
					Type: "presence_update",
					Data: map[string]interface{}{
						"user_id": client.UserID.String(),
						"status":  "online",
					},
				})
				h.sendToServer(client.ServerID, msg)
			}

			// Send the new client the full presence list
			presenceMsg, _ := json.Marshal(Message{
				Type: "presence_list",
				Data: map[string]interface{}{
					"user_ids": h.onlineUserIDs(client.ServerID),
				},
			})
			select {
			case client.Send <- presenceMsg:
			default:
			}

		case client := <-h.Unregister:
			if _, ok := h.Clients[client.ServerID]; ok {
				if _, exists := h.Clients[client.ServerID][client]; exists {
					delete(h.Clients[client.ServerID], client)
					close(client.Send)

					if len(h.Clients[client.ServerID]) == 0 {
						delete(h.Clients, client.ServerID)
					}

					log.Printf("Client unregistered. Server: %s", client.ServerID)

					// If user has no remaining connections, broadcast offline
					if !h.userConnectedToServer(client.UserID, client.ServerID) {
						msg, _ := json.Marshal(Message{
							Type: "presence_update",
							Data: map[string]interface{}{
								"user_id": client.UserID.String(),
								"status":  "offline",
							},
						})
						h.sendToServer(client.ServerID, msg)
					}
				}
			}

		case message := <-h.Broadcast:
			if serverClients, ok := h.Clients[message.ServerID]; ok {
				for client := range serverClients {
					select {
					case client.Send <- message.Message:
					default:
						close(client.Send)
						delete(h.Clients[message.ServerID], client)

						if len(h.Clients[message.ServerID]) == 0 {
							delete(h.Clients, message.ServerID)
						}
					}
				}
			}
		}
	}
}

func (h *Hub) BroadcastToServer(serverID gocql.UUID, message []byte) {
	h.Broadcast <- BroadcastMessage{
		ServerID: serverID,
		Message:  message,
	}
}
