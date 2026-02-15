package websocket

import (
	"log"
	"net/http"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"newcord/api/internal/middleware"
)

type WSHandler struct {
	Hub            *Hub
	AllowedOrigins []string
}

func NewWSHandler(hub *Hub, allowedOrigins []string) *WSHandler {
	return &WSHandler{Hub: hub, AllowedOrigins: allowedOrigins}
}

func (h *WSHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := gocql.ParseUUID(vars["server_id"])
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			for _, allowed := range h.AllowedOrigins {
				if origin == allowed {
					return true
				}
			}
			return false
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		Hub:      h.Hub,
		Conn:     conn,
		Send:     make(chan []byte, 1024),
		UserID:   userID,
		ServerID: serverID,
	}

	client.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
