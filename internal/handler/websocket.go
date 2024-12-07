package handler

import (
	"log"
	"net/http"
	"strings"

	"fearlessdraft-server/internal/service"
	"fearlessdraft-server/pkg/types"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type LobbyHandler struct {
	lobbyService *service.LobbyService
	upgrader     websocket.Upgrader
}

func NewLobbyHandler(ls *service.LobbyService) *LobbyHandler {
	return &LobbyHandler{
		lobbyService: ls,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (h *LobbyHandler) HandleLobbyWebSocket(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid lobby URL", http.StatusBadRequest)
		return
	}

	lobbyID := parts[2]
	roleStr := parts[3]

	lobby, exists := h.lobbyService.GetLobby(lobbyID)
	if !exists {
		http.Error(w, "Lobby not found", http.StatusNotFound)
		return
	}

	var role types.LobbyRole
	switch roleStr {
	case "blue":
		role = types.RoleBlueTeam
	case "red":
		role = types.RoleRedTeam
	case "spectator":
		role = types.RoleSpectator
	default:
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	User := &types.User{
		ID:   generateUserID(),
		Conn: conn,
		Role: role,
	}

	lobby.AddUser(User)

	h.handleUserConnection(lobby, User)
}

func (h *LobbyHandler) handleUserConnection(lobby *types.Lobby, User *types.User) {
	for {
		_, message, err := User.Conn.ReadMessage()
		if err != nil {
			h.removeUser(lobby, User)
			break
		}

		h.processMessage(lobby, User, message)
	}
}

func (h *LobbyHandler) processMessage(lobby *types.Lobby, User *types.User, message []byte) {
	// Implement role-based message processing
	switch User.Role {
	case types.RoleBlueTeam:
		// Blue team specific actions
	case types.RoleRedTeam:
		// Red team specific actions
	case types.RoleSpectator:
		// Spectator can only receive updates
	}
}

func (h *LobbyHandler) removeUser(lobby *types.Lobby, User *types.User) {
	lobby.RemoveUser(User.ID)
}

func generateUserID() string {
	return uuid.New().String()
}
