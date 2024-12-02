package main

import (
	"fmt"
	"log"
	"net/http"

	"fearlessdraft-server/internal/handler"
	"fearlessdraft-server/internal/service"
)

func main() {
	dataURL := "https://cdn.merakianalytics.com/riot/lol/resources/latest/en-US/championrates.json"

	championRatesService := service.NewChampionRatesService(dataURL)

	championRatesHandler := handler.NewChampionRatesHandler(championRatesService)

	http.HandleFunc("/proxy/championrates", championRatesHandler.HandleChampionRates)

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
