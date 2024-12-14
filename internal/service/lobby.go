package service

import (
	"fmt"
	"sync"

	"fearlessdraft-server/pkg/types"
)

type LobbyService struct {
	lobbies      map[string]*types.Lobby
	lobbiesMutex sync.RWMutex
}

type LobbyCreateResponse struct {
	LobbyID      string `json:"lobbyId"`
	BlueTeamURL  string `json:"blueTeamUrl"`
	RedTeamURL   string `json:"redTeamUrl"`
	SpectatorURL string `json:"spectatorUrl"`
}

func NewLobbyService() *LobbyService {
	return &LobbyService{
		lobbies: make(map[string]*types.Lobby),
	}
}

func (s *LobbyService) CreateLobby(options *types.DraftOptions, blueTeamName string, redTeamName string, champions []*types.DraftChampion) *LobbyCreateResponse {
	s.lobbiesMutex.Lock()
	defer s.lobbiesMutex.Unlock()

	lobby := types.NewLobby(*options, blueTeamName, redTeamName, champions)

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
