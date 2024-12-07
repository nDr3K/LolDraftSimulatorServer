package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"fearlessdraft-server/internal/handler"
	"fearlessdraft-server/internal/service"
	"fearlessdraft-server/pkg/types"
)

func main() {

	handleChampionRates()
	handleLobby()

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleChampionRates() {
	dataURL := "https://cdn.merakianalytics.com/riot/lol/resources/latest/en-US/championrates.json"

	championRatesService := service.NewChampionRatesService(dataURL)

	championRatesHandler := handler.NewChampionRatesHandler(championRatesService)

	http.HandleFunc("/proxy/championrates", championRatesHandler.HandleChampionRates)
}

func handleLobby() {
	lobbyService := service.NewLobbyService()

	lobbyHandler := handler.NewLobbyHandler(lobbyService)

	http.HandleFunc("/api/lobby/create", func(w http.ResponseWriter, r *http.Request) {

		var lobbyRequest struct {
			BlueTeamName string              `json:"blueTeamName"`
			RedTeamName  string              `json:"redTeamName"`
			Options      *types.DraftOptions `json:"options"`
		}

		err := json.NewDecoder(r.Body).Decode(&lobbyRequest)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Now you can access the parsed data
		lobbyResponse := lobbyService.CreateLobby(
			lobbyRequest.Options,
			lobbyRequest.BlueTeamName,
			lobbyRequest.RedTeamName,
		)

		json.NewEncoder(w).Encode(lobbyResponse)
	})

	http.HandleFunc("/ws/lobby/", lobbyHandler.HandleLobbyWebSocket)
}
