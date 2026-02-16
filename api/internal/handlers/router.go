package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/time/rate"
	"newcord/api/internal/db"
	"newcord/api/internal/middleware"
	"newcord/api/internal/websocket"
)

func NewRouter(
	cassandraDB *db.CassandraDB,
	wsHub *websocket.Hub,
	jwtSecret string,
	allowedOrigins []string,
) *mux.Router {
	r := mux.NewRouter()

	userRepo := db.NewUserRepository(cassandraDB)
	serverRepo := db.NewServerRepository(cassandraDB)
	channelRepo := db.NewChannelRepository(cassandraDB)
	messageRepo := db.NewMessageRepository(cassandraDB)
	inviteRepo := db.NewInviteRepository(cassandraDB)

	authHandler := NewAuthHandler(userRepo, jwtSecret)
	userHandler := NewUserHandler(userRepo)
	serverHandler := NewServerHandler(serverRepo, channelRepo)
	channelHandler := NewChannelHandler(channelRepo, serverRepo)
	messageHandler := NewMessageHandler(messageRepo, channelRepo, serverRepo, wsHub)
	inviteHandler := NewInviteHandler(inviteRepo, serverRepo)
	wsHandler := websocket.NewWSHandler(wsHub, allowedOrigins)

	authMiddleware := middleware.NewAuthMiddleware(jwtSecret)
	globalRateLimiter := middleware.NewRateLimiter(rate.Limit(20), 40)
	authRateLimiter := middleware.NewRateLimiter(rate.Limit(5), 10)

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.CORSWithOrigins(allowedOrigins))
	r.Use(globalRateLimiter.Limit)

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}).Methods("GET")

	api := r.PathPrefix("/api/v1").Subrouter()

	// Auth routes with stricter rate limiting
	authRoutes := api.PathPrefix("/auth").Subrouter()
	authRoutes.Use(authRateLimiter.Limit)
	authRoutes.HandleFunc("/register", authHandler.Register).Methods("POST")
	authRoutes.HandleFunc("/login", authHandler.Login).Methods("POST")

	protected := api.PathPrefix("").Subrouter()
	protected.Use(authMiddleware.Authenticate)

	protected.HandleFunc("/users/{id}", userHandler.GetUser).Methods("GET")
	protected.HandleFunc("/users/{id}", userHandler.UpdateUser).Methods("PUT")
	protected.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")

	protected.HandleFunc("/users/@me/servers", serverHandler.GetMyServers).Methods("GET")

	protected.HandleFunc("/servers", serverHandler.CreateServer).Methods("POST")
	protected.HandleFunc("/servers/{id}", serverHandler.GetServer).Methods("GET")
	protected.HandleFunc("/servers/{id}", serverHandler.UpdateServer).Methods("PUT")
	protected.HandleFunc("/servers/{id}", serverHandler.DeleteServer).Methods("DELETE")
	protected.HandleFunc("/servers/{id}/members", serverHandler.GetMembers).Methods("GET")
	protected.HandleFunc("/servers/{id}/members", serverHandler.AddMember).Methods("POST")
	protected.HandleFunc("/servers/{id}/invites", inviteHandler.CreateInvite).Methods("POST")
	protected.HandleFunc("/servers/{id}/invites", inviteHandler.GetServerInvites).Methods("GET")
	protected.HandleFunc("/invites/{code}/join", inviteHandler.JoinViaInvite).Methods("POST")

	protected.HandleFunc("/channels", channelHandler.CreateChannel).Methods("POST")
	protected.HandleFunc("/channels/{id}", channelHandler.GetChannel).Methods("GET")
	protected.HandleFunc("/channels/{id}", channelHandler.UpdateChannel).Methods("PUT")
	protected.HandleFunc("/channels/{id}", channelHandler.DeleteChannel).Methods("DELETE")
	protected.HandleFunc("/servers/{server_id}/channels", channelHandler.GetServerChannels).Methods("GET")

	protected.HandleFunc("/channels/{channel_id}/messages", messageHandler.CreateMessage).Methods("POST")
	protected.HandleFunc("/channels/{channel_id}/messages", messageHandler.GetMessages).Methods("GET")
	protected.HandleFunc("/channels/{channel_id}/messages/{message_id}", messageHandler.UpdateMessage).Methods("PUT")
	protected.HandleFunc("/channels/{channel_id}/messages/{message_id}", messageHandler.DeleteMessage).Methods("DELETE")

	protected.HandleFunc("/ws/{server_id}", wsHandler.ServeWS)

	return r
}
