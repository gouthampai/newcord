package handlers

import (
	"crypto/rand"
	"encoding/json"
	"log"
	"math/big"
	"net/http"
	"time"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"newcord/api/internal/db"
	"newcord/api/internal/middleware"
	"newcord/api/internal/models"
	"newcord/api/internal/websocket"
)

const inviteCodeChars = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"

type InviteHandler struct {
	inviteRepo *db.InviteRepository
	serverRepo *db.ServerRepository
	wsHub      *websocket.Hub
}

func NewInviteHandler(inviteRepo *db.InviteRepository, serverRepo *db.ServerRepository, wsHub *websocket.Hub) *InviteHandler {
	return &InviteHandler{inviteRepo: inviteRepo, serverRepo: serverRepo, wsHub: wsHub}
}

func (h *InviteHandler) broadcastWS(serverID gocql.UUID, msgType string, data interface{}) {
	wsMsg := websocket.Message{
		Type:      msgType,
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

type CreateInviteRequest struct {
	MaxUses   int `json:"max_uses"`
	ExpiresIn int `json:"expires_in"` // seconds, 0 = never
}

func generateCode(length int) (string, error) {
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(inviteCodeChars))))
		if err != nil {
			return "", err
		}
		b[i] = inviteCodeChars[n.Int64()]
	}
	return string(b), nil
}

func (h *InviteHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
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

	var req CreateInviteRequest
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	}

	code, err := generateCode(8)
	if err != nil {
		http.Error(w, "Failed to generate invite code", http.StatusInternalServerError)
		return
	}

	invite := &models.Invite{
		ServerID:  serverID,
		Code:      code,
		CreatedBy: userID,
		MaxUses:   req.MaxUses,
		Uses:      0,
	}

	if req.ExpiresIn > 0 {
		invite.ExpiresAt = time.Now().Add(time.Duration(req.ExpiresIn) * time.Second)
	}

	if err := h.inviteRepo.Create(invite); err != nil {
		http.Error(w, "Failed to create invite", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(invite)
}

func (h *InviteHandler) JoinViaInvite(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	code := vars["code"]

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	invite, err := h.inviteRepo.GetByCode(code)
	if err != nil {
		http.Error(w, "Invalid invite code", http.StatusNotFound)
		return
	}

	// Check expiry
	if !invite.ExpiresAt.IsZero() && time.Now().After(invite.ExpiresAt) {
		http.Error(w, "Invite has expired", http.StatusGone)
		return
	}

	// Check max uses
	if invite.MaxUses > 0 && invite.Uses >= invite.MaxUses {
		http.Error(w, "Invite has reached maximum uses", http.StatusGone)
		return
	}

	// Check if already a member
	if _, err := h.serverRepo.GetMember(invite.ServerID, userID); err == nil {
		http.Error(w, "Already a member of this server", http.StatusConflict)
		return
	}

	// Add as member
	member := &models.Member{
		ServerID: invite.ServerID,
		UserID:   userID,
		Role:     "member",
	}
	if err := h.serverRepo.AddMember(member); err != nil {
		http.Error(w, "Failed to join server", http.StatusInternalServerError)
		return
	}

	h.broadcastWS(invite.ServerID, "member_join", map[string]string{
		"server_id": invite.ServerID.String(),
		"user_id":   userID.String(),
	})

	// Increment uses
	_ = h.inviteRepo.IncrementUses(invite.ID, invite.Uses)

	// Return the server
	server, err := h.serverRepo.GetByID(invite.ServerID)
	if err != nil {
		http.Error(w, "Failed to get server info", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(server)
}

func (h *InviteHandler) GetServerInvites(w http.ResponseWriter, r *http.Request) {
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

	invites, err := h.inviteRepo.GetByServer(serverID)
	if err != nil {
		http.Error(w, "Failed to get invites", http.StatusInternalServerError)
		return
	}

	if invites == nil {
		invites = []*models.Invite{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invites)
}
