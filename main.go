package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func championRatesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	resp, err := http.Get("https://cdn.merakianalytics.com/riot/lol/resources/latest/en-US/championrates.json")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch champion rates: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read response body: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func main() {
	http.HandleFunc("/proxy/championrates", championRatesHandler)

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
