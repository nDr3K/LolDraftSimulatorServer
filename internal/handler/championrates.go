package handler

import (
	"encoding/json"
	"net/http"

	"fearlessdraft-server/internal/service"
)

type ChampionRatesHandler struct {
	service *service.ChampionRatesService
}

func NewChampionRatesHandler(s *service.ChampionRatesService) *ChampionRatesHandler {
	return &ChampionRatesHandler{
		service: s,
	}
}

func (h *ChampionRatesHandler) HandleChampionRates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	remappedData, err := h.service.FetchAndTransformRates()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	remappedJSON, err := json.Marshal(remappedData)
	if err != nil {
		http.Error(w, "Failed to generate JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(remappedJSON)
}
