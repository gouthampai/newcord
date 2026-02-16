package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gocql/gocql"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	ContextKeyUserID   contextKey = "user_id"
	ContextKeyUsername contextKey = "username"
)

type AuthMiddleware struct {
	jwtSecret []byte
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: []byte(jwtSecret)}
}

// GetUserID safely extracts the authenticated user ID from the request context.
func GetUserID(r *http.Request) (gocql.UUID, bool) {
	val := r.Context().Value(ContextKeyUserID)
	if val == nil {
		return gocql.UUID{}, false
	}
	uid, ok := val.(gocql.UUID)
	return uid, ok
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}
			tokenString = parts[1]
		} else if qToken := r.URL.Query().Get("token"); qToken != "" {
			// Fallback for WebSocket connections which cannot set headers
			tokenString = qToken
		} else {
			http.Error(w, "Missing authorization", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return m.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
			return
		}

		userID, err := gocql.ParseUUID(userIDStr)
		if err != nil {
			http.Error(w, "Invalid user ID format", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ContextKeyUserID, userID)
		ctx = context.WithValue(ctx, ContextKeyUsername, claims["username"])

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
