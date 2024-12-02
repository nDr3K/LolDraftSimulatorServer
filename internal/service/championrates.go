package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"fearlessdraft-server/pkg/types"
)

type ChampionRatesService struct {
	dataURL string
}

func NewChampionRatesService(url string) *ChampionRatesService {
	return &ChampionRatesService{
		dataURL: url,
	}
}

func (s *ChampionRatesService) MapRole(role string) string {
	switch role {
	case "TOP":
		return "top"
	case "JUNGLE":
		return "jungle"
	case "MIDDLE":
		return "mid"
	case "BOTTOM":
		return "bot"
	case "UTILITY":
		return "support"
	default:
		log.Printf("Warning: Invalid role '%s'", role)
		return role
	}
}

func (s *ChampionRatesService) FetchAndTransformRates() (*types.RemappedChampionRates, error) {
	resp, err := http.Get(s.dataURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch champion rates: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var originalData types.OriginalChampionRates
	if err := json.Unmarshal(body, &originalData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	remappedData := &types.RemappedChampionRates{
		Data: make(map[string]map[string]struct {
			PlayRate float64 `json:"playRate"`
		}),
	}

	for champID, roles := range originalData.Data {
		remappedData.Data[champID] = make(map[string]struct {
			PlayRate float64 `json:"playRate"`
		})

		for originalRole, roleData := range roles {
			newRole := s.MapRole(originalRole)
			remappedData.Data[champID][newRole] = roleData
		}
	}

	return remappedData, nil
}
