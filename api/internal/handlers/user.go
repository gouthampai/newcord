package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"newcord/api/internal/db"
	"newcord/api/internal/middleware"
)

type UserHandler struct {
	userRepo *db.UserRepository
}

func NewUserHandler(userRepo *db.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

type UpdateUserRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
	Status      *string `json:"status,omitempty"`
	Bio         *string `json:"bio,omitempty"`
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := gocql.ParseUUID(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	targetUserID, err := gocql.ParseUUID(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	authenticatedUserID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if targetUserID != authenticatedUserID {
		http.Error(w, "Forbidden: cannot modify other users", http.StatusForbidden)
		return
	}

	var updateReq UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetByID(targetUserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if updateReq.DisplayName != nil {
		if len(*updateReq.DisplayName) > 64 {
			http.Error(w, "Display name must be at most 64 characters", http.StatusBadRequest)
			return
		}
		user.DisplayName = *updateReq.DisplayName
	}
	if updateReq.AvatarURL != nil {
		user.AvatarURL = *updateReq.AvatarURL
	}
	if updateReq.Status != nil {
		status := *updateReq.Status
		if status != "online" && status != "offline" && status != "away" && status != "dnd" {
			http.Error(w, "Invalid status value", http.StatusBadRequest)
			return
		}
		user.Status = status
	}
	if updateReq.Bio != nil {
		if len(*updateReq.Bio) > 256 {
			http.Error(w, "Bio must be at most 256 characters", http.StatusBadRequest)
			return
		}
		user.Bio = *updateReq.Bio
	}

	if err := h.userRepo.Update(user); err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	targetUserID, err := gocql.ParseUUID(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	authenticatedUserID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if targetUserID != authenticatedUserID {
		http.Error(w, "Forbidden: cannot delete other users", http.StatusForbidden)
		return
	}

	if err := h.userRepo.Delete(targetUserID); err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
