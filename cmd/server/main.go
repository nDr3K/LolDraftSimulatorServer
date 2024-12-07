package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"fearlessdraft-server/cmd/server/middleware"
	"fearlessdraft-server/internal/handler"
	"fearlessdraft-server/internal/service"
	"fearlessdraft-server/pkg/types"
)

func main() {

	mux := http.NewServeMux()

	handleChampionRates(mux)
	handleLobby(mux)

	fmt.Println("Server starting on :8080")
	handlerMiddleware := middleware.CorsMiddleware(mux)
	log.Fatal(http.ListenAndServe(":8080", handlerMiddleware))
}

func handleChampionRates(mux *http.ServeMux) {
	dataURL := "https://cdn.merakianalytics.com/riot/lol/resources/latest/en-US/championrates.json"

	championRatesService := service.NewChampionRatesService(dataURL)

	championRatesHandler := handler.NewChampionRatesHandler(championRatesService)

	mux.HandleFunc("/proxy/championrates", championRatesHandler.HandleChampionRates)
}

func handleLobby(mux *http.ServeMux) {
	lobbyService := service.NewLobbyService()

	lobbyHandler := handler.NewLobbyHandler(lobbyService)

	mux.HandleFunc("/api/lobby/create", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")

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

		lobbyResponse := lobbyService.CreateLobby(
			lobbyRequest.Options,
			lobbyRequest.BlueTeamName,
			lobbyRequest.RedTeamName,
		)

		json.NewEncoder(w).Encode(lobbyResponse)
	})

	mux.HandleFunc("/ws/lobby/", lobbyHandler.HandleLobbyWebSocket)
}
