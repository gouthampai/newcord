package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"newcord/api/internal/db"
	"newcord/api/internal/models"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

type AuthHandler struct {
	userRepo  *db.UserRepository
	jwtSecret []byte
}

func NewAuthHandler(userRepo *db.UserRepository, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
	}
}

type RegisterRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

func (req *RegisterRequest) Validate() string {
	if len(req.Username) < 3 || len(req.Username) > 32 {
		return "username must be 3-32 characters"
	}
	if !usernameRegex.MatchString(req.Username) {
		return "username can only contain letters, numbers, and underscores"
	}
	if !emailRegex.MatchString(req.Email) {
		return "invalid email format"
	}
	if len(req.Password) < 8 {
		return "password must be at least 8 characters"
	}
	if len(req.Password) > 128 {
		return "password must be at most 128 characters"
	}
	if len(req.DisplayName) > 64 {
		return "display name must be at most 64 characters"
	}
	return ""
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if msg := req.Validate(); msg != "" {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// Check uniqueness
	if existing, _ := h.userRepo.GetByEmail(req.Email); existing != nil {
		http.Error(w, "Email already registered", http.StatusConflict)
		return
	}
	if existing, _ := h.userRepo.GetByUsername(req.Username); existing != nil {
		http.Error(w, "Username already taken", http.StatusConflict)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to process registration", http.StatusInternalServerError)
		return
	}

	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		DisplayName:  req.DisplayName,
		Status:       "offline",
	}

	if err := h.userRepo.Create(user); err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	token, err := h.generateToken(user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AuthResponse{
		Token: token,
		User:  user,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := h.generateToken(user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		Token: token,
		User:  user,
	})
}

func (h *AuthHandler) generateToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID.String(),
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24 * 7).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.jwtSecret)
}
