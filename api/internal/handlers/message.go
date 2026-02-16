package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"newcord/api/internal/db"
	"newcord/api/internal/middleware"
	"newcord/api/internal/models"
	"newcord/api/internal/websocket"
)

type MessageHandler struct {
	messageRepo *db.MessageRepository
	channelRepo *db.ChannelRepository
	serverRepo  *db.ServerRepository
	wsHub       *websocket.Hub
}

func NewMessageHandler(messageRepo *db.MessageRepository, channelRepo *db.ChannelRepository, serverRepo *db.ServerRepository, wsHub *websocket.Hub) *MessageHandler {
	return &MessageHandler{messageRepo: messageRepo, channelRepo: channelRepo, serverRepo: serverRepo, wsHub: wsHub}
}

func (h *MessageHandler) broadcastWS(serverID gocql.UUID, msgType string, channelID string, data interface{}) {
	wsMsg := websocket.Message{
		Type:      msgType,
		ChannelID: channelID,
		ServerID:  serverID.String(),
		Data:      data,
		Timestamp: time.Now(),
	}
	payload, err := json.Marshal(wsMsg)
	if err != nil {
		log.Printf("error marshaling WS broadcast: %v", err)
		return
	}
	h.wsHub.BroadcastToServer(serverID, payload)
}

type CreateMessageRequest struct {
	Content     string             `json:"content"`
	Type        models.MessageType `json:"type"`
	Attachments []string           `json:"attachments"`
}

func (h *MessageHandler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID, err := gocql.ParseUUID(vars["channel_id"])
	if err != nil {
		http.Error(w, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	var req CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Content) == 0 {
		http.Error(w, "Message content is required", http.StatusBadRequest)
		return
	}
	if len(req.Content) > 2000 {
		http.Error(w, "Message too long (max 2000 characters)", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Verify user is a member of the channel's server
	channel, err := h.channelRepo.GetByID(channelID)
	if err != nil {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}
	if _, err := h.serverRepo.GetMember(channel.ServerID, userID); err != nil {
		http.Error(w, "Forbidden: not a server member", http.StatusForbidden)
		return
	}

	message := &models.Message{
		ChannelID:   channelID,
		UserID:      userID,
		Content:     req.Content,
		Type:        req.Type,
		Attachments: req.Attachments,
	}

	if err := h.messageRepo.Create(message); err != nil {
		http.Error(w, "Failed to create message", http.StatusInternalServerError)
		return
	}

	h.broadcastWS(channel.ServerID, "message_create", channelID.String(), message)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(message)
}

func (h *MessageHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID, err := gocql.ParseUUID(vars["channel_id"])
	if err != nil {
		http.Error(w, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Verify user is a member of the channel's server
	channel, err := h.channelRepo.GetByID(channelID)
	if err != nil {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}
	if _, err := h.serverRepo.GetMember(channel.ServerID, userID); err != nil {
		http.Error(w, "Forbidden: not a server member", http.StatusForbidden)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 1
	}

	messages, err := h.messageRepo.GetByChannel(channelID, limit)
	if err != nil {
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func (h *MessageHandler) UpdateMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID, err := gocql.ParseUUID(vars["channel_id"])
	if err != nil {
		http.Error(w, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	messageID, err := gocql.ParseUUID(vars["message_id"])
	if err != nil {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	var updateReq struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(updateReq.Content) == 0 {
		http.Error(w, "Message content is required", http.StatusBadRequest)
		return
	}
	if len(updateReq.Content) > 2000 {
		http.Error(w, "Message too long (max 2000 characters)", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	message, err := h.messageRepo.GetByID(channelID, messageID)
	if err != nil {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	// Only the message author can edit it
	if message.UserID != userID {
		http.Error(w, "Forbidden: can only edit your own messages", http.StatusForbidden)
		return
	}

	message.Content = updateReq.Content

	if err := h.messageRepo.Update(message); err != nil {
		http.Error(w, "Failed to update message", http.StatusInternalServerError)
		return
	}

	channel, chErr := h.channelRepo.GetByID(channelID)
	if chErr == nil {
		h.broadcastWS(channel.ServerID, "message_update", channelID.String(), message)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(message)
}

func (h *MessageHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID, err := gocql.ParseUUID(vars["channel_id"])
	if err != nil {
		http.Error(w, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	messageID, err := gocql.ParseUUID(vars["message_id"])
	if err != nil {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	message, err := h.messageRepo.GetByID(channelID, messageID)
	if err != nil {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	// Author can delete their own messages. Server owner/admin can delete any message.
	if message.UserID != userID {
		channel, err := h.channelRepo.GetByID(channelID)
		if err != nil {
			http.Error(w, "Channel not found", http.StatusNotFound)
			return
		}
		member, err := h.serverRepo.GetMember(channel.ServerID, userID)
		if err != nil || (member.Role != "owner" && member.Role != "admin" && member.Role != "moderator") {
			http.Error(w, "Forbidden: cannot delete other users' messages", http.StatusForbidden)
			return
		}
	}

	if err := h.messageRepo.Delete(channelID, message.CreatedAt, messageID); err != nil {
		http.Error(w, "Failed to delete message", http.StatusInternalServerError)
		return
	}

	channel, chErr := h.channelRepo.GetByID(channelID)
	if chErr == nil {
		h.broadcastWS(channel.ServerID, "message_delete", channelID.String(), map[string]string{
			"id":         messageID.String(),
			"channel_id": channelID.String(),
		})
	}

	w.WriteHeader(http.StatusNoContent)
}
