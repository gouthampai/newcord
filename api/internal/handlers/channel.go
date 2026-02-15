package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"newcord/api/internal/db"
	"newcord/api/internal/middleware"
	"newcord/api/internal/models"
)

type ChannelHandler struct {
	channelRepo *db.ChannelRepository
	serverRepo  *db.ServerRepository
}

func NewChannelHandler(channelRepo *db.ChannelRepository, serverRepo *db.ServerRepository) *ChannelHandler {
	return &ChannelHandler{channelRepo: channelRepo, serverRepo: serverRepo}
}

type CreateChannelRequest struct {
	ServerID    string            `json:"server_id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        models.ChannelType `json:"type"`
	Position    int               `json:"position"`
}

func (h *ChannelHandler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	var req CreateChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Name) < 1 || len(req.Name) > 100 {
		http.Error(w, "Channel name must be 1-100 characters", http.StatusBadRequest)
		return
	}
	if len(req.Description) > 1024 {
		http.Error(w, "Description must be at most 1024 characters", http.StatusBadRequest)
		return
	}

	serverID, err := gocql.ParseUUID(req.ServerID)
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Verify user is owner or admin of the server
	member, err := h.serverRepo.GetMember(serverID, userID)
	if err != nil || (member.Role != "owner" && member.Role != "admin") {
		http.Error(w, "Forbidden: insufficient permissions to create channels", http.StatusForbidden)
		return
	}

	channel := &models.Channel{
		ServerID:    serverID,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Position:    req.Position,
	}

	if err := h.channelRepo.Create(channel); err != nil {
		http.Error(w, "Failed to create channel", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(channel)
}

func (h *ChannelHandler) GetChannel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID, err := gocql.ParseUUID(vars["id"])
	if err != nil {
		http.Error(w, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	channel, err := h.channelRepo.GetByID(channelID)
	if err != nil {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	// Verify user is a member of the server
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if _, err := h.serverRepo.GetMember(channel.ServerID, userID); err != nil {
		http.Error(w, "Forbidden: not a server member", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(channel)
}

func (h *ChannelHandler) GetServerChannels(w http.ResponseWriter, r *http.Request) {
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

	// Verify user is a member of the server
	if _, err := h.serverRepo.GetMember(serverID, userID); err != nil {
		http.Error(w, "Forbidden: not a server member", http.StatusForbidden)
		return
	}

	channels, err := h.channelRepo.GetByServer(serverID)
	if err != nil {
		http.Error(w, "Failed to get channels", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(channels)
}

func (h *ChannelHandler) UpdateChannel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID, err := gocql.ParseUUID(vars["id"])
	if err != nil {
		http.Error(w, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	channel, err := h.channelRepo.GetByID(channelID)
	if err != nil {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	// Verify user is owner or admin
	member, err := h.serverRepo.GetMember(channel.ServerID, userID)
	if err != nil || (member.Role != "owner" && member.Role != "admin") {
		http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
		return
	}

	var updateReq models.Channel
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if updateReq.Name != "" {
		if len(updateReq.Name) > 100 {
			http.Error(w, "Channel name must be at most 100 characters", http.StatusBadRequest)
			return
		}
		channel.Name = updateReq.Name
	}
	if updateReq.Description != "" {
		if len(updateReq.Description) > 1024 {
			http.Error(w, "Description must be at most 1024 characters", http.StatusBadRequest)
			return
		}
		channel.Description = updateReq.Description
	}
	if updateReq.Position != 0 {
		channel.Position = updateReq.Position
	}

	if err := h.channelRepo.Update(channel); err != nil {
		http.Error(w, "Failed to update channel", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(channel)
}

func (h *ChannelHandler) DeleteChannel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID, err := gocql.ParseUUID(vars["id"])
	if err != nil {
		http.Error(w, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	channel, err := h.channelRepo.GetByID(channelID)
	if err != nil {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	// Only owner or admin can delete channels
	member, err := h.serverRepo.GetMember(channel.ServerID, userID)
	if err != nil || (member.Role != "owner" && member.Role != "admin") {
		http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
		return
	}

	if err := h.channelRepo.Delete(channelID); err != nil {
		http.Error(w, "Failed to delete channel", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
