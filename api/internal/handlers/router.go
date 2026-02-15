package handlers

import (
	"github.com/gorilla/mux"
	"newcord/api/internal/db"
	"newcord/api/internal/middleware"
	"newcord/api/internal/websocket"
)

func NewRouter(
	cassandraDB *db.CassandraDB,
	wsHub *websocket.Hub,
	jwtSecret string,
) *mux.Router {
	r := mux.NewRouter()

	userRepo := db.NewUserRepository(cassandraDB)
	serverRepo := db.NewServerRepository(cassandraDB)
	channelRepo := db.NewChannelRepository(cassandraDB)
	messageRepo := db.NewMessageRepository(cassandraDB)

	authHandler := NewAuthHandler(userRepo, jwtSecret)
	userHandler := NewUserHandler(userRepo)
	serverHandler := NewServerHandler(serverRepo)
	channelHandler := NewChannelHandler(channelRepo)
	messageHandler := NewMessageHandler(messageRepo)
	wsHandler := websocket.NewWSHandler(wsHub)

	authMiddleware := middleware.NewAuthMiddleware(jwtSecret)

	r.Use(middleware.Logger)
	r.Use(middleware.CORS)

	api := r.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	api.HandleFunc("/auth/login", authHandler.Login).Methods("POST")

	protected := api.PathPrefix("").Subrouter()
	protected.Use(authMiddleware.Authenticate)

	protected.HandleFunc("/users/{id}", userHandler.GetUser).Methods("GET")
	protected.HandleFunc("/users/{id}", userHandler.UpdateUser).Methods("PUT")
	protected.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")

	protected.HandleFunc("/servers", serverHandler.CreateServer).Methods("POST")
	protected.HandleFunc("/servers/{id}", serverHandler.GetServer).Methods("GET")
	protected.HandleFunc("/servers/{id}", serverHandler.UpdateServer).Methods("PUT")
	protected.HandleFunc("/servers/{id}", serverHandler.DeleteServer).Methods("DELETE")
	protected.HandleFunc("/servers/{id}/members", serverHandler.GetMembers).Methods("GET")
	protected.HandleFunc("/servers/{id}/members", serverHandler.AddMember).Methods("POST")

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
