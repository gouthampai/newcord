package websocket

import (
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

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			if h.Clients[client.ServerID] == nil {
				h.Clients[client.ServerID] = make(map[*Client]bool)
			}
			h.Clients[client.ServerID][client] = true
			log.Printf("Client registered. Server: %s, Total clients: %d",
				client.ServerID, len(h.Clients[client.ServerID]))

		case client := <-h.Unregister:
			if _, ok := h.Clients[client.ServerID]; ok {
				if _, exists := h.Clients[client.ServerID][client]; exists {
					delete(h.Clients[client.ServerID], client)
					close(client.Send)

					if len(h.Clients[client.ServerID]) == 0 {
						delete(h.Clients, client.ServerID)
					}

					log.Printf("Client unregistered. Server: %s", client.ServerID)
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
