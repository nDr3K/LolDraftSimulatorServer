package service

import (
	"fmt"
	"sync"
	"time"

	"fearlessdraft-server/pkg/types"
)

type LobbyService struct {
	lobbies      map[string]*types.Lobby
	lobbiesMutex sync.RWMutex
	lobbyTimeout time.Duration
}

type LobbyCreateResponse struct {
	LobbyID      string `json:"lobbyId"`
	BlueTeamURL  string `json:"blueTeamUrl"`
	RedTeamURL   string `json:"redTeamUrl"`
	SpectatorURL string `json:"spectatorUrl"`
}

func NewLobbyService() *LobbyService {
	service := &LobbyService{
		lobbies:      make(map[string]*types.Lobby),
		lobbyTimeout: 5 * time.Minute,
	}
	go service.cleanupLobbies()
	return service
}

func (s *LobbyService) cleanupLobbies() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.lobbiesMutex.Lock()
		for lobbyID, lobby := range s.lobbies {
			if s.isLobbyInactive(lobby) {
				delete(s.lobbies, lobbyID)
				fmt.Printf("Removed inactive lobby: %s\n", lobbyID)
			}
		}
		s.lobbiesMutex.Unlock()
	}
}

func (s *LobbyService) isLobbyInactive(lobby *types.Lobby) bool {
	lobby.Mutex.RLock()
	defer lobby.Mutex.RUnlock()

	return len(lobby.Users) == 0
}

func (s *LobbyService) CreateLobby(
	options *types.DraftOptions,
	blueTeamName string,
	redTeamName string,
	champions []*types.DraftChampion,
	disabledChampionIds []*string,
) *LobbyCreateResponse {
	s.lobbiesMutex.Lock()
	defer s.lobbiesMutex.Unlock()

	lobby := types.NewLobby(*options, blueTeamName, redTeamName, champions, disabledChampionIds)

	s.lobbies[lobby.ID] = lobby

	baseURL := "/draft"
	return &LobbyCreateResponse{
		LobbyID:      lobby.ID,
		BlueTeamURL:  fmt.Sprintf("%s/%s/blue", baseURL, lobby.ID),
		RedTeamURL:   fmt.Sprintf("%s/%s/red", baseURL, lobby.ID),
		SpectatorURL: fmt.Sprintf("%s/%s/spectator", baseURL, lobby.ID),
	}
}

func (s *LobbyService) GetLobby(lobbyID string) (*types.Lobby, bool) {
	s.lobbiesMutex.RLock()
	defer s.lobbiesMutex.RUnlock()

	lobby, exists := s.lobbies[lobbyID]
	return lobby, exists
}
