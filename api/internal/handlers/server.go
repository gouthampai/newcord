package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"newcord/api/internal/db"
	"newcord/api/internal/middleware"
	"newcord/api/internal/models"
	"newcord/api/internal/websocket"
)

type ServerHandler struct {
	serverRepo  *db.ServerRepository
	channelRepo *db.ChannelRepository
	wsHub       *websocket.Hub
}

func NewServerHandler(serverRepo *db.ServerRepository, channelRepo *db.ChannelRepository, wsHub *websocket.Hub) *ServerHandler {
	return &ServerHandler{serverRepo: serverRepo, channelRepo: channelRepo, wsHub: wsHub}
}

func (h *ServerHandler) broadcastWS(serverID gocql.UUID, msgType string, data interface{}) {
	wsMsg := websocket.Message{
		Type:     msgType,
		ServerID: serverID.String(),
		Data:     data,
		Timestamp: time.Now(),
	}
	payload, err := json.Marshal(wsMsg)
	if err != nil {
		log.Printf("error marshaling WS broadcast: %v", err)
		return
	}
	h.wsHub.BroadcastToServer(serverID, payload)
}

type CreateServerRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IconURL     string `json:"icon_url"`
}

func (h *ServerHandler) CreateServer(w http.ResponseWriter, r *http.Request) {
	var req CreateServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Name) < 1 || len(req.Name) > 100 {
		http.Error(w, "Server name must be 1-100 characters", http.StatusBadRequest)
		return
	}
	if len(req.Description) > 1024 {
		http.Error(w, "Description must be at most 1024 characters", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	server := &models.Server{
		Name:        req.Name,
		Description: req.Description,
		IconURL:     req.IconURL,
		OwnerID:     userID,
	}

	if err := h.serverRepo.Create(server); err != nil {
		http.Error(w, "Failed to create server", http.StatusInternalServerError)
		return
	}

	member := &models.Member{
		ServerID: server.ID,
		UserID:   userID,
		Role:     "owner",
	}
	if err := h.serverRepo.AddMember(member); err != nil {
		log.Printf("Warning: failed to add owner as member for server %s: %v", server.ID, err)
	}

	defaultChannels := []models.Channel{
		{ServerID: server.ID, Name: "general", Type: "text", Position: 0},
		{ServerID: server.ID, Name: "voice-chat", Type: "voice", Position: 1},
	}
	for i := range defaultChannels {
		if err := h.channelRepo.Create(&defaultChannels[i]); err != nil {
			log.Printf("Warning: failed to create default channel %q for server %s: %v", defaultChannels[i].Name, server.ID, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(server)
}

func (h *ServerHandler) GetServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := gocql.ParseUUID(vars["id"])
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	server, err := h.serverRepo.GetByID(serverID)
	if err != nil {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(server)
}

func (h *ServerHandler) UpdateServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := gocql.ParseUUID(vars["id"])
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user is owner or admin
	member, err := h.serverRepo.GetMember(serverID, userID)
	if err != nil || (member.Role != "owner" && member.Role != "admin") {
		http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
		return
	}

	var updateReq CreateServerRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	server, err := h.serverRepo.GetByID(serverID)
	if err != nil {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	if updateReq.Name != "" {
		if len(updateReq.Name) > 100 {
			http.Error(w, "Server name must be at most 100 characters", http.StatusBadRequest)
			return
		}
		server.Name = updateReq.Name
	}
	if updateReq.Description != "" {
		if len(updateReq.Description) > 1024 {
			http.Error(w, "Description must be at most 1024 characters", http.StatusBadRequest)
			return
		}
		server.Description = updateReq.Description
	}
	if updateReq.IconURL != "" {
		server.IconURL = updateReq.IconURL
	}

	if err := h.serverRepo.Update(server); err != nil {
		http.Error(w, "Failed to update server", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(server)
}

func (h *ServerHandler) DeleteServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := gocql.ParseUUID(vars["id"])
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Only owner can delete a server
	server, err := h.serverRepo.GetByID(serverID)
	if err != nil {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}
	if server.OwnerID != userID {
		http.Error(w, "Forbidden: only the server owner can delete it", http.StatusForbidden)
		return
	}

	// Broadcast before delete — clients are still connected to the server's hub room
	h.broadcastWS(serverID, "server_delete", map[string]string{
		"server_id": serverID.String(),
	})

	if err := h.serverRepo.Delete(serverID); err != nil {
		http.Error(w, "Failed to delete server", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ServerHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := gocql.ParseUUID(vars["id"])
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Verify requester is a member
	if _, err := h.serverRepo.GetMember(serverID, userID); err != nil {
		http.Error(w, "Forbidden: not a server member", http.StatusForbidden)
		return
	}

	members, err := h.serverRepo.GetMembers(serverID)
	if err != nil {
		http.Error(w, "Failed to get members", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

func (h *ServerHandler) GetMyServers(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	servers, err := h.serverRepo.GetServersByUser(userID)
	if err != nil {
		http.Error(w, "Failed to get servers", http.StatusInternalServerError)
		return
	}

	if servers == nil {
		servers = []*models.Server{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(servers)
}

func (h *ServerHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := gocql.ParseUUID(vars["id"])
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Verify requester is owner or admin to add members
	requester, err := h.serverRepo.GetMember(serverID, userID)
	if err != nil || (requester.Role != "owner" && requester.Role != "admin") {
		http.Error(w, "Forbidden: insufficient permissions to add members", http.StatusForbidden)
		return
	}

	var member models.Member
	if err := json.NewDecoder(r.Body).Decode(&member); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	member.ServerID = serverID
	if member.Role == "" {
		member.Role = "member"
	}

	if err := h.serverRepo.AddMember(&member); err != nil {
		http.Error(w, "Failed to add member", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(member)
}
