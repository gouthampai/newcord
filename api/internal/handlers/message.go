package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"newcord/api/internal/db"
	"newcord/api/internal/models"
)

type MessageHandler struct {
	messageRepo *db.MessageRepository
}

func NewMessageHandler(messageRepo *db.MessageRepository) *MessageHandler {
	return &MessageHandler{messageRepo: messageRepo}
}

type CreateMessageRequest struct {
	Content     string            `json:"content"`
	Type        models.MessageType `json:"type"`
	Attachments []string          `json:"attachments"`
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

	userID := r.Context().Value("user_id").(gocql.UUID)

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

	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
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

	message, err := h.messageRepo.GetByID(channelID, messageID)
	if err != nil {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	message.Content = updateReq.Content

	if err := h.messageRepo.Update(message); err != nil {
		http.Error(w, "Failed to update message", http.StatusInternalServerError)
		return
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

	message, err := h.messageRepo.GetByID(channelID, messageID)
	if err != nil {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	if err := h.messageRepo.Delete(channelID, message.CreatedAt, messageID); err != nil {
		http.Error(w, "Failed to delete message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
