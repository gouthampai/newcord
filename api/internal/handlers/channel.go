package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"newcord/api/internal/db"
	"newcord/api/internal/models"
)

type ChannelHandler struct {
	channelRepo *db.ChannelRepository
}

func NewChannelHandler(channelRepo *db.ChannelRepository) *ChannelHandler {
	return &ChannelHandler{channelRepo: channelRepo}
}

type CreateChannelRequest struct {
	ServerID    string                `json:"server_id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Type        models.ChannelType    `json:"type"`
	Position    int                   `json:"position"`
}

func (h *ChannelHandler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	var req CreateChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	serverID, err := gocql.ParseUUID(req.ServerID)
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
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

	var updateReq models.Channel
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	channel, err := h.channelRepo.GetByID(channelID)
	if err != nil {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	if updateReq.Name != "" {
		channel.Name = updateReq.Name
	}
	if updateReq.Description != "" {
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

	if err := h.channelRepo.Delete(channelID); err != nil {
		http.Error(w, "Failed to delete channel", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
