package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"newcord/api/internal/db"
	"newcord/api/internal/models"
)

type ServerHandler struct {
	serverRepo *db.ServerRepository
}

func NewServerHandler(serverRepo *db.ServerRepository) *ServerHandler {
	return &ServerHandler{serverRepo: serverRepo}
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

	userID := r.Context().Value("user_id").(gocql.UUID)

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
	h.serverRepo.AddMember(member)

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

	var updateReq models.Server
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
		server.Name = updateReq.Name
	}
	if updateReq.Description != "" {
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

	members, err := h.serverRepo.GetMembers(serverID)
	if err != nil {
		http.Error(w, "Failed to get members", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

func (h *ServerHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := gocql.ParseUUID(vars["id"])
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
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
